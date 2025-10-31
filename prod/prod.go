package prod

import "strings"

var isProd bool = false

func SetProdByEnv(env string) {
	if strings.HasPrefix(env, "staging") || strings.HasPrefix(env, "prod") {
		isProd = true
	} else {
		isProd = false
	}
}

func SetProd(prod bool) {
	isProd = prod
}

func IsProd() bool {
	return isProd
}

func IsStagingProd(env string) bool {
	if env == "staging" || env == "prod" {
		return true
	}
	return false
}
