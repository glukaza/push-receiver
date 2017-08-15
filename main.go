package main

import (
	"fmt"
	"strconv"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"log"
	"os"
	"os/exec"
	"io/ioutil"
	"strings"
	"encoding/json"
	"github.com/go-redis/redis"
	"bytes"
	"bufio"
	"io"
	"reflect"
)

var client *redis.Client

type RequestBody struct {
	Repository 	Repository
	After		string
	Project_id	int
	User_email	string
	Ref 		string
}

type Repository struct {
	Name string
}

type ProjectConfig struct {
	Name     	string
	Job		string
	Host    	string
	RemoveJob 	string
	Branches	[]string
}

type configure struct {
	Project		[]ProjectConfig
}

type UserInfo struct {
	Skype 		string
}

func main() {
	router := httprouter.New()
	router.GET("/", Health)
	router.POST("/", Index)
	router.GET("/get-commits/:project", GetCommits)
	router.GET("/save-commits/", SaveCommits)

	log.Println(http.ListenAndServe(":8081", router))
}

func checkConnectToRedis() {
	pong, err := connectToRedis()
	for err != nil {
		pong, err = connectToRedis()
	}
	fmt.Println(pong, err)
}

func connectToRedis() (string, error) {
	var host = os.Getenv("PU_REDIS_HOST")
	var port = os.Getenv("PU_REDIS_PORT")
	client = redis.NewClient(&redis.Options{
		Addr:     host+":"+port,
		Password: "",
		DB:       0,
	})
	return client.Ping().Result()
}

func GetCommits(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	checkConnectToRedis()

	var project = ps.ByName("project")
	val, err := client.LRange(project, -100, 100).Result()
	check(err)
	for _, value := range val {
		fmt.Fprint(w, value)
	}

	_, err = client.Del(project).Result()
	check(err)

	client.Close()
}

func SaveCommits(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	checkConnectToRedis()

	var project = r.URL.Query().Get("project")
	var number = r.URL.Query().Get("number")
	var checkUrl = strings.Replace(r.URL.Query().Get("url"), " ", "%20", -1) + number + "/api/xml?wrapper=changes&xpath=//changeSet//item//msg"

	out, _ := exec.Command("bash", "-c", "curl -u " + os.Getenv("PU_JENKINS_USER") + ":" + os.Getenv("PU_JENKINS_USER_TOKEN") + " -s '" + checkUrl + "' | grep -oP '(?<=[\\#]).([A-Z-_a-z0-9]+)(?=[\\:])' | tr [a-z] [A-Z]").Output()
	//out, _ := exec.Command("bash", "-c", "curl -u aandronov:fad8aaf0ac96d8b9999a99a4943c8204 -s 'http://jenkins.cis.local/view/www.veeam.com/job/veeam-deploy/14908/api/xml?wrapper=changes&xpath=//changeSet//item//msg' | grep -oP '(?<=[\\#]).([A-Z-_a-z0-9]+)(?=[\\:])' | tr [a-z] [A-Z]").Output()

	readbuffer := bytes.NewBuffer([]byte(out))
	reader := bufio.NewReader(readbuffer)

	for {
		buffer, err := reader.ReadBytes('\n')
		var value = string(buffer)

		result := client.LPush(project, value).Err()
		check(result);

		if err != nil {
			break
		}

		if (err == io.EOF) {
			break
		}
	}

	client.Close()
}

func Health(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprint(w, "OK")
}

func Index(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var dataObject RequestBody
	var globalConfig []configure
	var project ProjectConfig

	var client http.Client

	fmt.Println("++++++++START++++++++")

	body, err := ioutil.ReadAll(r.Body)
	check(err)

	err = json.Unmarshal(body, &dataObject)
	check(err)

	globalConfig = getConfig("/etc/push-receivers/global.json")
	branch := strings.Split(dataObject.Ref, "/")[2]
	skypeUser := getSkypeUser(dataObject.User_email)

	for l := range globalConfig[0].Project {
		if (dataObject.Repository.Name == globalConfig[0].Project[l].Name) {
			project = globalConfig[0].Project[l]
		}
	}

	if project.IsEmpty() {
		fmt.Println(" +++ not project for " + dataObject.Repository.Name + " in config +++")
		fmt.Println("++++++++FINISH++++++++")
		return
	}

	if(dataObject.Repository.Name == "Forum" && branch == "vermilion") {
		project.Job  = "forum4.veeam.com";
		project.Host = "forum4";
	}

	if (dataObject.After == "0000000000000000000000000000000000000000") {
		project.Job = project.RemoveJob
	}

	fmt.Println(dataObject.Repository.Name + " +++ PROCESSING JOB: " + project.Job + " +++")

	req, _ := http.NewRequest("GET", os.Getenv("PU_JENKINS_URL"), nil)

	q := req.URL.Query()
	q.Add("job", project.Job)
	q.Add("token", os.Getenv("PU_JENKINS_TOKEN"))
	q.Add("delay", "0sec")
	q.Add("BRANCH_TO_BUILD", branch)
	q.Add("TARGET_SERVER", project.Host)
	q.Add("SKYPE", skypeUser)
	q.Add("HEAD_COMMIT", dataObject.After)
	q.Add("PROJECT_ID", strconv.Itoa(dataObject.Project_id))
	q.Add("ASSETS_ENV", "demo")

	req.URL.RawQuery = q.Encode()
//	fmt.Println(req.URL.RawQuery)

	if project.Branches != nil {
		branchExist := false
		for b := range project.Branches {
			if project.Branches[b] == branch {
				branchExist = true
			}
		}

		if(!branchExist) {
			fmt.Println("Reject " + dataObject.Repository.Name + " not branch '" + branch + "' build")
			fmt.Println("++++++++FINISH++++++++")
			return
		}
	}

	resp, _ := client.Do(req)
	fmt.Println(dataObject.Repository.Name + " +++ Jenkins status code: " + strconv.FormatInt(int64(resp.StatusCode), 10) + " +++")
	fmt.Println("++++++++FINISH++++++++")
}

func check(e error) {
	if e != nil {
		fmt.Println(e)
	}
}

func getConfig(file string) []configure {
	var configData []configure
	if _, err := os.Stat(file); err == nil {
		project, err := ioutil.ReadFile(file)
		err = json.Unmarshal(project, &configData)
		check(err)
	}
	return configData
}

func getSkypeUser(email string) string {
	var userInfo []UserInfo
	args := []string{"--header", "PRIVATE-TOKEN: " + os.Getenv("PU_SKYPE_GITLAB_TOKEN"), os.Getenv("PU_SKYPE_GITLAB") + email}

	if out, err := exec.Command("curl", args...).Output(); err == nil {
		json.Unmarshal(out, &userInfo)
		return string(userInfo[0].Skype)
	}

	return "glukolov3"
}

func (config ProjectConfig) IsEmpty() bool {
	return reflect.DeepEqual(config,ProjectConfig{})
}
