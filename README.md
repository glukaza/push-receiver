#### Push reciver
##### Integration between gitlab and jenkins

##### Use in Docker or in k8s
1. For use in k8s - need rewrite manifests in k8s/push-receivers
1. For use Docker you can just ```make compile && make docker``` (GOPATH Installed before)

##### Dependencies
1. Redis

##### How It Works
1. go to you gitlab and setup webhook or system hook to push-receiver url
2. go to jenkins and create job, what receiver is started and add Token for triggered remotly
3. if you need - you can setup another job for clearing remotely
4. setup Push Receiver config:
   1. "name": "projectName in Gitlab",
   1. "job": "JobNameInJenkins",
   1. "host": "HostParametersInJenkins(not required)",
   1. "removeJob": "JobNameForCleanUp(not required)",
   1. "branches": ["master", "test"]
5. setup Push Receiver Env Variables:
   1. PU_JENKINS_URL - Jenkins Url (http://jenkins.cis.local/buildByToken/buildWithParameter)
   2. PU_JENKINS_TOKEN - Token for triggering job
   3. PU_JENKINS_USER - Jenkins user, who can get information from jenkins
   4. PU_JENKINS_USER_TOKEN - This user token
   5. PU_SKYPE_GITLAB - gitlab api url (https://gitlab/api/v3/users?search=)
   6. PU_SKYPE_GITLAB_TOKEN - token for gitlab connect
   7. PU_REDIS_HOST - redis host
   8. PU_REDIS_PORT - redis port

   