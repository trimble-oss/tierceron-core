package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"bitbucket.org/dexterchaney/whoville/utils"
	"bitbucket.org/dexterchaney/whoville/vaulthelper/kv"
	pb "bitbucket.org/dexterchaney/whoville/webapi/rpc/apinator"
	gql "github.com/graphql-go/graphql"
)

//Currently selected environments
var SelectedEnvironment []string

// Currently selected init environments
var SelectedInitEnvironment []string

// Server implements the twirp api server endpoints
type Server struct {
	VaultToken          string
	VaultAddr           string
	VaultAPITokenSecret []byte
	GQLSchema           gql.Schema
	Log                 *log.Logger
}

// NewServer Creates a new server struct and initializes the GraphQL schema
func NewServer(VaultAddr string, VaultToken string) *Server {
	s := Server{}
	s.VaultToken = VaultToken
	s.VaultAddr = VaultAddr
	s.Log = log.New(os.Stdout, "[INFO]", log.LstdFlags)
	s.VaultAPITokenSecret = nil

	return &s
}

// InitConfig initializes configuration information for the server.
func (s *Server) InitConfig(env string) error {
	connInfo, err := s.GetConfig(env, "apiLogins/meta")
	if err != nil {
		utils.LogErrorObject(err, s.Log, false)
		return err
	}
	vaultAPITokenSecretString, ok := connInfo["vaultApiTokenSecret"].(string)
	if !ok {
		err := fmt.Errorf("Missing vaultApiTokenSecret")
		utils.LogErrorObject(err, s.Log, false)
		return err
	}

	s.VaultAPITokenSecret = []byte(vaultAPITokenSecretString)
	return nil
}

// ListServiceTemplates lists the templates under the requested project
func (s *Server) ListServiceTemplates(ctx context.Context, req *pb.ListReq) (*pb.ListResp, error) {
	mod, err := kv.NewModifier(s.VaultToken, s.VaultAddr)
	if err != nil {
		utils.LogErrorObject(err, s.Log, false)
		return nil, err
	}

	listPath := "templates/" + req.Project + "/" + req.Service
	secret, err := mod.List(listPath)
	if err != nil {
		utils.LogErrorObject(err, s.Log, false)
		return nil, err
	}
	if secret == nil {
		err := fmt.Errorf("Could not find any templates under %s", req.Project+"/"+req.Service)
		utils.LogErrorObject(err, s.Log, false)
		return nil, err
	}
	utils.LogWarningsObject(secret.Warnings, s.Log, false)
	if len(secret.Warnings) > 0 {
		err := errors.New("Warnings generated from vault " + req.Project + "/" + req.Service)
		utils.LogErrorObject(err, s.Log, false)
		return nil, err
	}

	fileNames := []string{}
	for _, fileName := range secret.Data["keys"].([]interface{}) {
		if strFile, ok := fileName.(string); ok {
			if strFile[len(strFile)-1] != '/' { // Skip subdirectories where template files are stored
				fileNames = append(fileNames, strFile)
			}
		}
	}

	return &pb.ListResp{
		Templates: fileNames,
	}, nil
}

// GetTemplate makes a request to the vault for the template found in <project>/<service>/<file>/template-file
// Returns the template data in base64 and the template's extension. Returns any errors generated by vault
func (s *Server) GetTemplate(ctx context.Context, req *pb.TemplateReq) (*pb.TemplateResp, error) {
	// Connect to the vault
	mod, err := kv.NewModifier(s.VaultToken, s.VaultAddr)
	if err != nil {
		utils.LogErrorObject(err, s.Log, false)
		return nil, err
	}

	// Get template data from information in request.
	path := "templates/" + req.Project + "/" + req.Service + "/" + req.File + "/template-file"
	data, err := mod.ReadData(path)
	if err != nil {
		utils.LogErrorObject(err, s.Log, false)
		return nil, err
	}
	if data == nil {
		err := errors.New("No file " + req.File + " under " + req.Project + "/" + req.Service)
		utils.LogErrorObject(err, s.Log, false)
		return nil, err
	}

	// Return retrieved data in response
	return &pb.TemplateResp{
		Data: data["data"].(string),
		Ext:  data["ext"].(string)}, nil
}

// Validate checks the vault to see if the requested credentials are validated
func (s *Server) Validate(ctx context.Context, req *pb.ValidationReq) (*pb.ValidationResp, error) {
	mod, err := kv.NewModifier(s.VaultToken, s.VaultAddr)
	if err != nil {
		utils.LogErrorObject(err, s.Log, false)
		return nil, err
	}
	mod.Env = req.Env

	servicePath := "verification/" + req.Project + "/" + req.Service
	data, err := mod.ReadData(servicePath)
	if err != nil {
		utils.LogErrorObject(err, s.Log, false)
		return nil, err
	}

	if data == nil {
		err := errors.New("No verification for " + req.Project + "/" + req.Service + " found under " + req.Env + " environment")
		utils.LogErrorObject(err, s.Log, false)
		return nil, err
	}
	return &pb.ValidationResp{IsValid: data["verified"].(bool)}, nil
}

//GetValues gets values requested from the vault
func (s *Server) GetValues(ctx context.Context, req *pb.GetValuesReq) (*pb.ValuesRes, error) {
	mod, err := kv.NewModifier(s.VaultToken, s.VaultAddr)
	if err != nil {
		utils.LogErrorObject(err, s.Log, false)
		return nil, err
	}
	environments := []*pb.ValuesRes_Env{}
	envStrings := SelectedEnvironment
	//Only display staging in prod mode
	for i, other := range envStrings {
		if other == "prod" {
			envStrings = append(envStrings[:i], envStrings[i+1:]...)
			break
		}
	}
	for _, e := range envStrings {
		mod.Env = "local/" + e
		userPaths, err := mod.List("values/")
		if err != nil {
			return nil, err
		}
		if userPaths == nil {
			continue
		}
		if localEnvs, ok := userPaths.Data["keys"].([]interface{}); ok {
			for _, env := range localEnvs {
				envStrings = append(envStrings, strings.Trim("local/"+e+"/"+env.(string), "/"))
			}
		}
	}
	for _, environment := range envStrings {
		mod, err := kv.NewModifier(s.VaultToken, s.VaultAddr)
		if err != nil {
			utils.LogErrorObject(err, s.Log, false)
			return nil, err
		}
		mod.Env = environment
		projects := []*pb.ValuesRes_Env_Project{}
		//get a list of projects under values
		projectPaths, err := s.getPaths(mod, "values/")
		if err != nil {
			utils.LogErrorObject(err, s.Log, false)
			return nil, err
		}

		for _, projectPath := range projectPaths {
			services := []*pb.ValuesRes_Env_Project_Service{}
			//get a list of files under project
			servicePaths, err := s.getPaths(mod, projectPath)
			//fmt.Println("servicePaths")
			//fmt.Println(servicePaths)
			if err != nil {
				utils.LogErrorObject(err, s.Log, false)
				return nil, err
			}

			for _, servicePath := range servicePaths {
				files := []*pb.ValuesRes_Env_Project_Service_File{}
				//get a list of files under project
				filePaths, err := s.getPaths(mod, servicePath)
				if err != nil {
					utils.LogErrorObject(err, s.Log, false)
					return nil, err
				}

				for _, filePath := range filePaths {
					vals := []*pb.ValuesRes_Env_Project_Service_File_Value{}
					//get a list of values
					valueMap, err := mod.ReadData(filePath)
					if err != nil {
						err := fmt.Errorf("Unable to fetch data from %s", filePath)
						utils.LogErrorObject(err, s.Log, false)
						return nil, err
					}
					if valueMap != nil {

						for key, value := range valueMap {
							kv := &pb.ValuesRes_Env_Project_Service_File_Value{Key: key, Value: value.(string), Source: "value"}
							vals = append(vals, kv)
							//data = append(data, value.(string))
							//fmt.Println(value)
						}

					}
					if len(vals) > 0 {
						file := &pb.ValuesRes_Env_Project_Service_File{Name: getPathEnd(filePath), Values: vals}
						files = append(files, file)
					}
				}
				if len(files) > 0 {
					service := &pb.ValuesRes_Env_Project_Service{Name: getPathEnd(servicePath), Files: files}
					services = append(services, service)
				}
			}
			if len(services) > 0 {
				project := &pb.ValuesRes_Env_Project{Name: getPathEnd(projectPath), Services: services}
				projects = append(projects, project)
			}
		}
		if len(projects) > 0 {
			env := &pb.ValuesRes_Env{Name: environment, Projects: projects}
			environments = append(environments, env)
		}
	}
	return &pb.ValuesRes{
		Envs: environments,
	}, nil
}
func (s *Server) getPaths(mod *kv.Modifier, pathName string) ([]string, error) {
	secrets, err := mod.List(pathName)
	//fmt.Println("secrets " + pathName)
	//fmt.Println(secrets)
	pathList := []string{}
	if err != nil {
		utils.LogErrorObject(err, s.Log, false)
		return nil, fmt.Errorf("Unable to list paths under %s in %s", pathName, mod.Env)
	} else if secrets != nil {
		//add paths
		slicey := secrets.Data["keys"].([]interface{})
		//fmt.Println("secrets are")
		//fmt.Println(slicey)
		for _, pathEnd := range slicey {
			//List is returning both pathEnd and pathEnd/
			path := pathName + pathEnd.(string)
			pathList = append(pathList, path)
		}
		//fmt.Println("pathList")
		//fmt.Println(pathList)
		return pathList, nil
	}
	return pathList, nil
}
func (s *Server) getTemplateFilePaths(mod *kv.Modifier, pathName string) ([]string, error) {
	secrets, err := mod.List(pathName)
	pathList := []string{}
	if err != nil {
		utils.LogErrorObject(err, s.Log, false)
		return nil, fmt.Errorf("Unable to list paths under %s in %s", pathName, mod.Env)
	} else if secrets != nil {
		//add paths
		slicey := secrets.Data["keys"].([]interface{})

		for _, pathEnd := range slicey {
			//List is returning both pathEnd and pathEnd/
			path := pathName + pathEnd.(string)
			pathList = append(pathList, path)
		}

		subPathList := []string{}
		for _, path := range pathList {
			subsubList, _ := s.templateFileRecurse(mod, path)
			for _, subsub := range subsubList {
				//List is returning both pathEnd and pathEnd/
				subPathList = append(subPathList, subsub)
			}
		}
		if len(subPathList) != 0 {
			return subPathList, nil
		}
	}
	return pathList, nil
}
func (s *Server) templateFileRecurse(mod *kv.Modifier, pathName string) ([]string, error) {
	subPathList := []string{}
	subsecrets, err := mod.List(pathName)
	if err != nil {
		utils.LogErrorObject(err, s.Log, false)
		return subPathList, err
	} else if subsecrets != nil {
		subslice := subsecrets.Data["keys"].([]interface{})
		if subslice[0] != "template-file" {
			for _, pathEnd := range subslice {
				//List is returning both pathEnd and pathEnd/
				subpath := pathName + pathEnd.(string)
				subsublist, _ := s.templateFileRecurse(mod, subpath)
				if len(subsublist) != 0 {
					for _, subsub := range subsublist {
						//List is returning both pathEnd and pathEnd/
						subPathList = append(subPathList, subsub)
					}
				}
				subPathList = append(subPathList, subpath)
			}
		} else {
			subPathList = append(subPathList, pathName)
		}
	}
	return subPathList, nil
}

func getPathEnd(path string) string {
	strs := strings.Split(path, "/")
	for strs[len(strs)-1] == "" {
		strs = strs[:len(strs)-1]
	}
	return strs[len(strs)-1]
}

// UpdateAPI takes the passed URL and downloads the given build of the UI
func (s *Server) UpdateAPI(ctx context.Context, req *pb.UpdateAPIReq) (*pb.NoParams, error) {
	scriptPath := "./getArtifacts.sh"
	//buildNum := strconv.FormatInt(int64(req.Build), 10)
	buildNum := req.Build
	//fmt.Println(buildNum)
	for len(buildNum) < 5 {
		buildNum = "0" + buildNum
	}
	cmd := exec.Command(scriptPath, buildNum)
	cmd.Dir = "/etc/opt/vaultAPI"
	err := cmd.Run()
	utils.LogErrorObject(err, s.Log, false)
	return &pb.NoParams{}, err
}

// ResetServer resets vault token.
func (s *Server) ResetServer(ctx context.Context, req *pb.ResetReq) (*pb.NoParams, error) {
	if s.VaultToken == "" {
		s.VaultToken = req.PrivToken
	}

	if s.VaultAPITokenSecret == nil {
		var targetEnv string
		for _, e := range SelectedEnvironment {
			targetEnv = e
			if e == "dev" {
				break
			} else if e == "staging" {
				break
			}
		}
		s.InitConfig(targetEnv)
	}
	return &pb.NoParams{}, nil
}

// CheckConnection checks the server connection
func (s *Server) CheckConnection(ctx context.Context, req *pb.NoParams) (*pb.CheckConnResp, error) {
	if len(s.VaultToken) == 0 {
		return &pb.CheckConnResp{
			Connected: false,
		}, nil
	}
	return &pb.CheckConnResp{
		Connected: true,
	}, nil
}

// Environments selects environments based on dev or production mode
func (s *Server) Environments(ctx context.Context, req *pb.NoParams) (*pb.EnvResp, error) {
	return &pb.EnvResp{
		Env: SelectedEnvironment,
	}, nil

}
