package container

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/yungkcx/mydocker/image"
	"github.com/yungkcx/mydocker/util"
)

// Info is info of container.
type Info struct {
	Name    string `json:"name"`
	Command string `json:"command"`
	Created string `json:"created"`
	Status  string `json:"status"`
	Pid     string `json:"pid"`
	Image   string `json:"image"`
}

// Container status and name of config file.
var (
	RUNNING    = "Running"
	EXITED     = "Exited"
	ConfigName = "config.json"
)

// NewContainer returns rootdir of the new container and error.
func NewContainer(imageName string) (string, error) {
	// Create container's directory and lower-id file.
	if exist, err := image.IsImageExist(imageName); !exist {
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("Image is no exist: %s", imageName)
	}
	containerName := newContainerName(imageName)
	containerRoot := filepath.Join(util.ContainersDir, containerName)
	if err := os.MkdirAll(containerRoot, 0700); err != nil {
		return "", fmt.Errorf("Error: %v", err)
	}
	lowerIDPath := filepath.Join(containerRoot, "lower-id")
	if err := ioutil.WriteFile(lowerIDPath, []byte(imageName), 0744); err != nil {
		return "", fmt.Errorf("Error write to %s: %v", lowerIDPath, err)
	}
	return containerRoot, nil
}

func newContainerName(imageName string) string {
	return strconv.FormatUint(util.BKDRHash(imageName+strconv.FormatInt(time.Now().Unix(), 16)), 16)
}

func getContainers() ([]string, error) {
	if exist, err := util.PathExist(util.ContainersDir); !exist {
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	f, err := os.Open(util.ContainersDir)
	if err != nil {
		return nil, fmt.Errorf("Error: %v", err)
	}
	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil, fmt.Errorf("Error read ContainersDir: %v", err)
	}
	return names, nil
}

// ListContainers list containers.
func ListContainers() error {
	names, err := getContainers()
	if err != nil {
		return err
	}

	// Read configs.
	var infos []*Info
	for _, name := range names {
		path := filepath.Join(util.ContainersDir, name, ConfigName)
		info, err := getInfo(path)
		if err != nil {
			return err
		}
		infos = append(infos, info)
	}

	// Print image directories.
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprintf(w, "NAME\tPID\tIMAGE\tCOMMAND\tCREATED\tSTATUS\n")
	for _, info := range infos {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			info.Name,
			info.Pid,
			info.Image,
			info.Command,
			info.Created,
			info.Status,
		)
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("Error flushing tabwriter: %v", err)
	}

	return nil
}

func getInfo(path string) (*Info, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error reading %s: %v", path, err)
	}
	var info Info
	if err := json.Unmarshal(content, &info); err != nil {
		return nil, fmt.Errorf("Error json.unmarshal: %v", err)
	}
	return &info, nil
}

func writeInfo(info *Info, path string) error {
	jsonBytes, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("Error json Marshal: %v", err)
	}
	if err := ioutil.WriteFile(path, jsonBytes, 0700); err != nil {
		return fmt.Errorf("Error WriteFile %s: %v", path, err)
	}
	return nil
}

// ExitInfo set config.json to {"status": "Exited"}.
func ExitInfo(root string) error {
	configPath := filepath.Join(root, ConfigName)
	info, err := getInfo(configPath)
	if err != nil {
		return err
	}
	info.Status = EXITED
	return writeInfo(info, configPath)
}

// SetConfigFile writes info of the container to the config.json file.
func SetConfigFile(image string, root string, cmdArray []string, pid int) error {
	name := path.Base(root)
	command := strings.Join(cmdArray, " ")
	created := time.Now().Format("2006-01-02 15:04:05")
	info := &Info{
		Name:    name,
		Pid:     strconv.Itoa(pid),
		Image:   image,
		Command: command,
		Created: created,
		Status:  RUNNING,
	}

	configPath := filepath.Join(root, ConfigName)
	if err := writeInfo(info, configPath); err != nil {
		return err
	}
	return nil
}

// RemoveContainer removes container.
func RemoveContainer(names ...string) error {
	for _, name := range names {
		dir := filepath.Join(util.ContainersDir, name)
		if exist, err := util.PathExist(dir); !exist {
			if err != nil {
				return err
			}
			continue
		}
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("Error remove container %s: %v", dir, err)
		}
		fmt.Println(name)
	}
	return nil
}
