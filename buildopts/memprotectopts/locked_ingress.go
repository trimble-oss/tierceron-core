package memprotectopts

import (
	"bytes"
	"io"
	"unsafe"
)

const lockedReadInitialSize = 4096

// LockedReadAll reads r to EOF directly into locked memory obtained up front,
// so the payload never lands in unprotected storage. Release the returned
// buffer with LockedFree once its sensitive content has been consumed.
func LockedReadAll(r io.Reader) ([]byte, error) {
	buf, err := LockedAlloc(lockedReadInitialSize)
	if err != nil {
		return nil, err
	}
	total := 0
	for {
		if total == len(buf) {
			grown, err := LockedAlloc(len(buf) * 2)
			if err != nil {
				LockedFree(buf)
				return nil, err
			}
			copy(grown, buf[:total])
			if err := LockedFree(buf); err != nil {
				LockedFree(grown)
				return nil, err
			}
			buf = grown
		}
		n, err := r.Read(buf[total:])
		total += n
		if err == io.EOF {
			return buf[:total], nil
		}
		if err != nil {
			LockedFree(buf)
			return nil, err
		}
	}
}

// ExtractLockedStrings copies the values of the named JSON string fields from
// body into locked memory and redacts them in body, so a later Unmarshal of
// body cannot copy them to unprotected storage. The returned strings are views
// into locked memory and live until MemUnprotectAll. Values containing escape
// sequences are left in body for the caller to protect after decoding.
func ExtractLockedStrings(body []byte, keys ...string) (map[string]string, error) {
	spans := make(map[string][2]int, len(keys))
	total := 0
	i := 0
	for i < len(body) {
		if body[i] != '"' {
			i++
			continue
		}
		keyStart := i + 1
		keyEnd, ok := scanJSONString(body, keyStart)
		if !ok {
			break
		}
		i = keyEnd + 1
		j := skipJSONSpace(body, i)
		if j >= len(body) || body[j] != ':' {
			continue
		}
		j = skipJSONSpace(body, j+1)
		if j >= len(body) || body[j] != '"' {
			continue
		}
		valStart := j + 1
		valEnd, ok := scanJSONString(body, valStart)
		if !ok {
			break
		}
		i = valEnd + 1
		value := body[valStart:valEnd]
		if len(value) == 0 || bytes.IndexByte(value, '\\') >= 0 {
			// Empty or escaped; the decoded form differs from the raw span.
			continue
		}
		for _, want := range keys {
			if string(body[keyStart:keyEnd]) == want {
				if _, seen := spans[want]; !seen {
					spans[want] = [2]int{valStart, valEnd}
					total += len(value)
				}
				break
			}
		}
	}
	extracted := make(map[string]string, len(spans))
	if total == 0 {
		return extracted, nil
	}
	locked, err := LockedAlloc(total)
	if err != nil {
		return nil, err
	}
	offset := 0
	for key, span := range spans {
		value := body[span[0]:span[1]]
		copy(locked[offset:], value)
		extracted[key] = unsafe.String(&locked[offset], len(value))
		offset += len(value)
		// Keep body valid JSON of identical shape while removing the secret.
		for p := span[0]; p < span[1]; p++ {
			body[p] = 'X'
		}
	}
	return extracted, nil
}

// scanJSONString returns the index of the closing quote of the JSON string
// whose content begins at i.
func scanJSONString(body []byte, i int) (int, bool) {
	for i < len(body) {
		switch body[i] {
		case '\\':
			i += 2
		case '"':
			return i, true
		default:
			i++
		}
	}
	return 0, false
}

func skipJSONSpace(body []byte, i int) int {
	for i < len(body) {
		switch body[i] {
		case ' ', '\t', '\n', '\r':
			i++
		default:
			return i
		}
	}
	return i
}

// ReleaseStringValues releases the protected string values of m and blanks
// them, so per-operation configuration does not accumulate locked memory.
// Values whose backing is shared with a map in shared are kept, as are
// pointer-typed entries, which share cached values by convention.
func ReleaseStringValues(m map[string]any, shared ...map[string]any) error {
	retained := make(map[*byte]struct{})
	for _, sharedMap := range shared {
		for _, value := range sharedMap {
			switch v := value.(type) {
			case string:
				if len(v) > 0 {
					retained[unsafe.StringData(v)] = struct{}{}
				}
			case *string:
				if v != nil && len(*v) > 0 {
					retained[unsafe.StringData(*v)] = struct{}{}
				}
			}
		}
	}
	var result error
	for key, value := range m {
		text, isString := value.(string)
		if !isString || len(text) == 0 {
			continue
		}
		if _, keep := retained[unsafe.StringData(text)]; keep {
			continue
		}
		if err := MemRelease(nil, &text); err != nil {
			result = err
			continue
		}
		m[key] = text
	}
	return result
}
