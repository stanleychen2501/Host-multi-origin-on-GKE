# Source IP detection application on GKE with SSL and behind CDN and serve multi domain

##Table of Contents
###Prerequisites
###File List
###1.Build go flie and Docker
###2.Bring your own SSL on GCP (Webber)
###3.Apply yaml on GKE
###4.Next let's enable CloudCDN on google console
###5.Verification

#### Source IP detection application -CDN,Kubenetes,Nginx-ingress,and Docker


##### <span style="font-style:italic; color:#ff8c00">***This lab is going to show how to build an application which is able to detect the source IP of the end user by reading the HTTP header, "X-Forworded-For". Since the HTTP packet is sent from the end user through the internet, the route is possible to be similar to below. How could the application get the real source IP correctly? Let;s do it together to accomplish this task step by step.***</span>

<img src="https://storage.googleapis.com/stanleymdbucket/md_image/detect-IP-GKE-CDN/detecet-IP-GKE-CDN-09.png" alt="Snow" color="black" style="width:1000px">


####Prerequisites:

| Required packages/sdk  | Description  | Memo                      |
|:------------- |:---------------:      | -------------:            |
| Docker        | docker sdk/pakages.   |Build docker image         |
| Kubectl       | Kubenetes sdk/pakages |Operate your K8s Cluster   |
| gcloud| GCP cloud shell sdk.  |Create GKE Cluster/Build and push docker image |

####File list
| File  | Description  |unit |
|:------------- |:---------------:  |:---------------:  |   
| main.go        | go code   | 1 |
| go.mod      | run go on your local device | 1 |
| go.sum| run go on your local device | 1 |
| Dockerfile| Dockerfile for each domain  | 2 |
| deployment.yaml| deployment YAML for each domain  | 2|
| service.yaml| service YAML for each domain  | 2|
| ingress.yaml| ingress YAML  | 1 |

##1. Build go flie and Docker
####1-1. Create main.go file(Note. If you have multi domain, please create file for each)

Jump to your directory then create your main.go.

<span style="font-style:italic; color:#ff8c00">***suggest create different directory for each domain***</span>

```
//Your domain(Identify which file)
//main go
package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/sebest/xff"
)

func main() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//ips := strings.Split(r.Header.Get("X-Forwarded-For"))
		xff := strings.Split(r.Header.Get("X-Forwarded-For"), ", ")
		log.Printf("xff: %+v", xff)
		w.Write([]byte("A website , XFF IP is " + strings.Join(xff, ", ") + "\n"))
	})
// "A website" is help us to identify which site.

	xffmw, _ := xff.Default()
	http.ListenAndServe(":8080", xffmw.Handler(handler))
}

```
####1-2. If you have a golang enviroment locally, you can build your go module and binary locally by below command(Optional)


`$ go mod init main go`

`$ go build -o xff`

go mod sample

```
module main.go

go 1.14

require github.com/sebest/xff v0.0.0-20160910043805-6c115e0ffa35
```
go sum smaple

```
github.com/sebest/xff v0.0.0-20160910043805-6c115e0ffa35 h1:eajwn6K3weW5cd1ZXLu2sJ4pvwlBiCWY4uDejOr73gM=
github.com/sebest/xff v0.0.0-20160910043805-6c115e0ffa35/go.mod h1:wozgYq9WEBQBaIJe4YZ0qTSFAMxmcwBhQH0fO0R34Z0=
```

####1-3. Create Dockerfile (Note. If you have multi domain, please create the Dockerfile for each)

Jump to the directory of your domain to create each file.

The first Dockerfile as below

```
#your first domain here
FROM golang:1.12-alpine

RUN apk add --no-cache git

# Set the Current Working Directory inside the container
WORKDIR /app/

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
#COPY go.sum .

RUN go mod download

COPY . .

# Build the Go app
RUN go build -o ./out/xff .


# This container exposes the port "Your port(mapping main.go file)" to the outside world
EXPOSE 8080

# Run the binary program produced by `go install`
CMD ["./out/xff"]
```

Jump to the directory of your domain to create each file.

The second Dockerfile as below


```
#your Second domain here
FROM golang:1.12-alpine

RUN apk add --no-cache git

# Set the Current Working Directory inside the container
WORKDIR /app/

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
#COPY go.sum .

RUN go mod download

COPY . .

# Build the Go app
RUN go build -o ./out/xff .


# This container exposes the port "Your port(mapping main.go file)" to the outside world
EXPOSE 8081

# Run the binary program produced by `go install`
CMD ["./out/xff"]
```
####1-4. Upload your docker image to the google container registry.Two ways you can do

Jump to the directory of your domain to run the docker command.

docker command:

```
# build docker image from your dokcer file, you have run the below command for your each docker image
docker build -t [YOUR-IMAGE-NAME] .


#tag your docker image for uploading to gcr.io or asia.gcr.io
*Note. Please do it for each docker image*
#[YOUR-PROJECT-NAME]
docker tag [YOUR-IMAGE-NAME] asia.gcr.io/[YOUR-PROJECT-NAME]/[YOUR-IMAGE-NAME]


#push your image to gcr.io or asia.gcr.io
*Note. Please do it for each docker image*
docker push asia.gcr.io/[YOUR-PROJECT-NAME]/[YOUR-IMAGE-NAME]


```
<span style="font-style:italic; color:#ff8c00">***Example:***</span>

```
#run docker build command as below.

$ docker build -t myxff .

#Then you will see the result as below.

Sending build context to Docker daemon  50.18kB
Step 1/9 : FROM golang:1.12-alpine
 ---> 76bddfb5e55e
Step 2/9 : RUN apk add --no-cache git
 ---> Using cache
 ---> 6f03ec9eed64
Step 3/9 : WORKDIR /app/
 ---> Running in 0bb359dc0996
Removing intermediate container 0bb359dc0996
 ---> 8b5d3346b5ff
Step 4/9 : COPY go.mod .
 ---> dd5a4df0febb
Step 5/9 : RUN go mod download
 ---> Running in 5d6f1df94779
go: finding github.com/sebest/xff v0.0.0-20160910043805-6c115e0ffa35
Removing intermediate container 5d6f1df94779
 ---> 75f72eaa8c3e
Step 6/9 : COPY . .
 ---> 32a8e933d78f
Step 7/9 : RUN go build -o ./out/xff .
 ---> Running in f155a20d22b3
Removing intermediate container f155a20d22b3
 ---> 4015aadba3af
Step 8/9 : EXPOSE 8080
 ---> Running in 0cb5c043b8bf
Removing intermediate container 0cb5c043b8bf
 ---> 63f12a20116b
Step 9/9 : CMD ["./out/xff"]
 ---> Running in ef1bbc2a0455
Removing intermediate container ef1bbc2a0455
 ---> fac5327d2dba
Successfully built fac5327d2dba
Successfully tagged myxff:latest
```

<span style="font-style:italic; color:red">*Run docker tag command.*</sapn>

`
$ docker tag myxff asia.gcr.io/stanleylab/myxff
`

<span style="font-style:italic; color:red">*There are not show any response, you can run <span style="font-style:italic; color:blue">"docker image ls"</span>*</sapn>


`
$ docker run image ls
`

<span style="font-style:italic; color:red">*You can see your image*</sapn>

```
REPOSITORY                                  TAG                 IMAGE ID            CREATED             SIZE
asia.gcr.io/stanleylab/myxff                latest              fac5327d2dba        6 minutes ago       370MB
```

<span style="font-style:italic; color:red">*Run docker push*</sapn>

`
$ docker push asia.gcr.io/stanleylab/myxff
`

<span style="font-style:italic; color:red">*You can see the result as below.*</sapn>

```
The push refers to repository [asia.gcr.io/stanleylab/myxff]
e25856fd12b4: Pushed
68daa66c8942: Pushed
17e244049ea9: Pushed
79d28e1109c6: Pushed
23cd76f94aaa: Pushed
8ed0f107989c: Layer already exists
7306dca01e79: Layer already exists
3957f7032fc4: Layer already exists
12c4e92b2d48: Layer already exists
45182158f5da: Layer already exists
5216338b40a7: Layer already exists
latest: digest: sha256:dcf9adf906dcc7a657812537936d9fc2233a99a0abb0d1396720d122d50160a5 size: 2618
```

gcloud build command:(Optional)

<span style="font-style:italic; color:#ff8c00">***Note. Please do it for each docker image***</span>

`
$ gcloud builds submit --tag asia.gcr.io/[YOUR-PROJECT-NAME]/[YOUR-IMAGE-NAME]
`

<span style="font-style:italic; color:#ff8c00">***Example:***</span>

`
$ gcloud builds submit --tag asia.gcr.io/stanleylab/myxff
`

<span style="font-style:italic; color:red">*You can see the result as below.*</span>

```
Creating temporary tarball archive of 19 file(s) totalling 32.8 KiB before compression.
Uploading tarball of [.] to [gs://stanleylab_cloudbuild/source/1603179843.955711-937a9542bba34be790f6abb164978acc.tgz]
Created [https://cloudbuild.googleapis.com/v1/projects/stanleylab/builds/5014bae1-ddd5-4bfc-a94e-90280b28975d].
Logs are available at [https://console.cloud.google.com/cloud-build/builds/5014bae1-ddd5-4bfc-a94e-90280b28975d?project=824476710360].
--------------------------------------------------------- REMOTE BUILD OUTPUT ---------------------------------------------------------
starting build "5014bae1-ddd5-4bfc-a94e-90280b28975d"
```

```
FETCHSOURCE
Fetching storage object: gs://stanleylab_cloudbuild/source/1603179843.955711-937a9542bba34be790f6abb164978acc.tgz#1603179845660922
Copying gs://stanleylab_cloudbuild/source/1603179843.955711-937a9542bba34be790f6abb164978acc.tgz#1603179845660922...
/ [1 files][ 11.3 KiB/ 11.3 KiB]
Operation completed over 1 objects/11.3 KiB.
BUILD
Already have image (with digest): gcr.io/cloud-builders/docker
Sending build context to Docker daemon  50.18kB
Step 1/9 : FROM golang:1.12-alpine
1.12-alpine: Pulling from library/golang
c9b1b535fdd9: Pulling fs layer
cbb0d8da1b30: Pulling fs layer
d909eff28200: Pulling fs layer
665fbbf998e4: Pulling fs layer
4985b1919860: Pulling fs layer
665fbbf998e4: Waiting
4985b1919860: Waiting
d909eff28200: Verifying Checksum
d909eff28200: Download complete
cbb0d8da1b30: Verifying Checksum
cbb0d8da1b30: Download complete
c9b1b535fdd9: Verifying Checksum
c9b1b535fdd9: Download complete
4985b1919860: Verifying Checksum
4985b1919860: Download complete
c9b1b535fdd9: Pull complete
cbb0d8da1b30: Pull complete
d909eff28200: Pull complete
665fbbf998e4: Verifying Checksum
665fbbf998e4: Download complete
665fbbf998e4: Pull complete
4985b1919860: Pull complete
Digest: sha256:3f8e3ad3e7c128d29ac3004ac8314967c5ddbfa5bfa7caa59b0de493fc01686a
Status: Downloaded newer image for golang:1.12-alpine
 ---> 76bddfb5e55e
Step 2/9 : RUN apk add --no-cache git
 ---> Running in 83b7cf91a32c
fetch http://dl-cdn.alpinelinux.org/alpine/v3.11/main/x86_64/APKINDEX.tar.gz
fetch http://dl-cdn.alpinelinux.org/alpine/v3.11/community/x86_64/APKINDEX.tar.gz
(1/5) Installing nghttp2-libs (1.40.0-r1)
(2/5) Installing libcurl (7.67.0-r1)
(3/5) Installing expat (2.2.9-r1)
(4/5) Installing pcre2 (10.34-r1)
(5/5) Installing git (2.24.3-r0)
Executing busybox-1.31.1-r9.trigger
OK: 22 MiB in 20 packages
Removing intermediate container 83b7cf91a32c
 ---> a614d632b062
Step 3/9 : WORKDIR /app/
 ---> Running in 95e27252c62f
Removing intermediate container 95e27252c62f
 ---> adb39ad19b76
Step 4/9 : COPY go.mod .
 ---> f05aec346848
Step 5/9 : RUN go mod download
 ---> Running in 5a932907c2a4
go: finding github.com/sebest/xff v0.0.0-20160910043805-6c115e0ffa35
Removing intermediate container 5a932907c2a4
 ---> 203cd8339dca
Step 6/9 : COPY . .
 ---> 109a849d04d7
Step 7/9 : RUN go build -o ./out/xff .
 ---> Running in 762ca0ed8128
Removing intermediate container 762ca0ed8128
 ---> 8c51d57ab9ba
Step 8/9 : EXPOSE 8080
 ---> Running in 6ea1b291240c
Removing intermediate container 6ea1b291240c
 ---> c74558363501
Step 9/9 : CMD ["./out/xff"]
 ---> Running in 7bd58a0e996d
Removing intermediate container 7bd58a0e996d
 ---> 090f74be8f3a
Successfully built 090f74be8f3a
Successfully tagged asia.gcr.io/stanleylab/myxff:latest
PUSH
Pushing asia.gcr.io/stanleylab/myxff
```

```
The push refers to repository [asia.gcr.io/stanleylab/myxff]
5ba06be00370: Preparing
705d718c302d: Preparing
4c12c329a93d: Preparing
ab44808f1eb6: Preparing
4b4363a260b8: Preparing
d0d527097126: Preparing
7306dca01e79: Preparing
3957f7032fc4: Preparing
12c4e92b2d48: Preparing
45182158f5da: Preparing
5216338b40a7: Preparing
d0d527097126: Waiting
7306dca01e79: Waiting
3957f7032fc4: Waiting
12c4e92b2d48: Waiting
45182158f5da: Waiting
5216338b40a7: Waiting
4b4363a260b8: Pushed
4c12c329a93d: Pushed
ab44808f1eb6: Pushed
7306dca01e79: Layer already exists
3957f7032fc4: Layer already exists
705d718c302d: Pushed
12c4e92b2d48: Layer already exists
45182158f5da: Layer already exists
5216338b40a7: Layer already exists
5ba06be00370: Pushed
d0d527097126: Pushed
latest: digest: sha256:941289698aed4f0783ccc34113a67ed3801ba4171ec6e9d5221693905cceb5d6 size: 2618
DONE
---------------------------------------------------------------------------------------------------------------------------------------

ID                                    CREATE_TIME                DURATION  SOURCE                                                                                    IMAGES                                  STATUS
5014bae1-ddd5-4bfc-a94e-90280b28975d  2020-10-20T07:44:07+00:00  37S       gs://stanleylab_cloudbuild/source/1603179843.955711-937a9542bba34be790f6abb164978acc.tgz  asia.gcr.io/stanleylab/myxff (+1 more)  SUCCESS


To take a quick anonymous survey, run:
  $ gcloud survey
```

##2. Set up your own SSL on GCP
####2-1.prerequisites for SSL settings

<span style="font-style:italic; color:red">*Private key encryption format and length: RSA 2048*. </span>
The maximum length for a certificate that you use with AWS CloudFront is 2048 bits, even though AWS ACM supports larger keys. Uploading a certificate to the AWS Identity and Access Management (IAM) certificate store: maximum size of the public key is 2048 bits.
GCP CloudCDN has the same requirement as AWS Cloudfront.

<span style="font-style:italic; color:red">*Certificate file format: pem*</span>

####2-2. Separately edit ingress.yaml for 2 domains: 
#####2-2-1 Create a Secret that holds your first certificate and key:


<span style="font-style:italic; color:#ff8c00">***[YOUR-FIRST-SECRET-NAME] will use in your "ingress.yaml"***</span>
`
$ kubectl create secret tls [YOUR-FIRST-SECRET-NAME] --cert [YOUR-FIRST-CERT-FILE] --key [YOUR-FIRST-KEY-FILE]
`

#####2-2-2 Create a Secret that holds your second certificate and key:


`
$ kubectl create secret tls [YOUR-FIRST-SECRET-NAME] --cert [YOUR-SECOND-CERT-FILE] --key [YOUR-SECOND-KEY-FILE]
`

Refer to the below sample command and output.

```
#sample command
$ kubectl create secret tls demo-secret --cert test_cn326_cn.pem --key test_cn326_cn.key

#output
secret/demo-secret created
```

##3. Apply yaml on GKE

####3-1. Create and Apply deployment the YAML file

In this lab, we suggest to implement pods separately and all the configuration files are also stored in different directory. You have to
store the YAML files in different directories. Or, feel free to contact us for further assistance.

<span style="font-style:italic; color:#ff8c00">***Note. If you have multi domain, please create file for each***</span>

deployment.yaml (use different port in different yaml)

```
apiVersion: apps/v1
kind: Deployment
#deployment name
metadata:    
  name: [YOUR-DEPLOYMENT-NAME]
spec:
  selector:
    matchLabels:
#Your app name      
      app: [YOUR-APP-NAME]
  replicas: 1
  template:
    metadata:
      labels:        
        app: [YOUR-APP-NAME]
    spec:
      containers:
       - name: [YOUR-CONTAINER-NAME]
         image: [YOUR-IMAGE]
#liveness, change the LB check path, the port have to mapping with Doockerfile and main.go
         livenessProbe: 
            httpGet:
              path: /xff
              port: [YOUR-SERVICE-PORT]
            initialDelaySeconds: 5
            periodSeconds: 5
#Your service port, that mapping with main.go and Dockfile           
         ports:
           - containerPort: [YOUR-SERVICE-PORT]
```

Refer to the below sample command and output.

```
#Sample command
$ kubectl apply -f deployment.yaml

#Outputs
deployment.apps/xff is created
```

`
$ kubectl get pods
`

<img src="https://storage.googleapis.com/stanleymdbucket/md_image/detect-IP-GKE-CDN/detecet-IP-GKE-CDN-01.jpg" alt="Snow" color="black" style="width:500px">

####3-2. Create Service YAML file and apply

You have to store the YAML files in different directories as well.

<span style="font-style:italic; color:#ff8c00">***Note. If you have multi domain, please create file for each***</span>

service.yaml (use different port in different yaml)

```
apiVersion: v1
kind: Service
metadata:
  name: [YOUR-SERVICE-NAME]
spec:
  type: NodePort
  selector:
    #mapping to deployment.yaml
    app: [YOUR-APP-NAME]
  ports:
  - name: [YOUR-PORT-NAME]
    protocol: TCP
    port: [YOUR-SERVICE-PORT]
    #mapping to deployment.yaml 
    targetPort: [YOUR-SERVICE-PORT]
```

Refer to the below sample command and output.

```
#Sample command
$ kubectl apply -f cn-service.yaml

#Outputs
service/xff is created
```

`
$ kubectl get service
`

<img src="https://storage.googleapis.com/stanleymdbucket/md_image/detect-IP-GKE-CDN/detecet-IP-GKE-CDN-02.jpg" alt="Snow" color="black" style="width:500px">

####3-3. Create ingress YAML file and apply

Jump to your working directory then create the ingress.yaml.

<span style="font-style:italic; color:#ff8c00">***Note. The ingress just need one***</span>

ingress.yaml

<span style="font-style:italic; color:#daa520">Kindly note. the </span><span style="font-style:italic; color:#ff8c00">***[YOUR-FIRST-PORT-NAME],  [YOUR-SECOND-PORT-NAME], [YOUR-FIRST-SERVICE-NAME] and [YOUR-SECOND-SERVICE-NAME]***</span><span style="font-style:italic; color:#daa520"> is refer to your each service.yaml file.</spn>


```
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: [YOUR-INGRESSS-NAME]
spec:
  tls:
  - secretName: [YOUR-FIRST-SECRET-NAME]
  - secretName: [YOUR-SECOND-SECRET-NAME]
  rules:
  - host: [YOUR-FIRST-DOMAIN]
    http:
       paths: 
        - backend:
            serviceName: [YOUR-FIRST-SERVICE-NAME]
            servicePort: [YOUR-FIRST-PORT-NAME]
  - host: [YOUR-SECOND-DOMAIN]
    http:
       paths: 
        - backend:
            serviceName: [YOUR-SECOND-SERVICE-NAME]
            servicePort: [YOUR-SECOND-PORT-NAME]
```

Refer to the below sample command and output.

```
#Sample command
$ kubectl apply -f cn-service.yaml

#Outputs
ingress/xff-ingress is created
```


`
$ kubectl get ingress
`

<img src="https://storage.googleapis.com/stanleymdbucket/md_image/detect-IP-GKE-CDN/detecet-IP-GKE-CDN-03.jpg" alt="Snow" color="black" style="width:500px">

Test your each URL https://YOUR-DOMAIN/

You can see your client IP will show on the browser

<img src="https://storage.googleapis.com/stanleymdbucket/md_image/detect-IP-GKE-CDN/detecet-IP-GKE-CDN-04.jpg" alt="Snow" color="black" style="width:1000px" border="2px">

And we can check the IP is correct.

<img src="https://storage.googleapis.com/stanleymdbucket/md_image/detect-IP-GKE-CDN/detecet-IP-GKE-CDN-10.jpg" alt="Snow" color="black" style="width:1000px" border="2px">

##4. Next let's enable CloudCDN on google console


Going to the console > jump to Load balance page the add your backend origin. 

<img src="https://storage.googleapis.com/stanleymdbucket/md_image/detect-IP-GKE-CDN/detecet-IP-GKE-CDN-05.jpg" alt="Snow" color="black" style="width:500px" border="2px">

<span style="font-style:italic; color:red">*Continue !!*</span>

<img src="https://storage.googleapis.com/stanleymdbucket/md_image/detect-IP-GKE-CDN/detecet-IP-GKE-CDN-06.jpg" alt="Snow" color="black" style="width:500px" border="2px">


Choose your load balancer as your origin(Your ingress)

<img src="https://storage.googleapis.com/stanleymdbucket/md_image/detect-IP-GKE-CDN/detecet-IP-GKE-CDN-07.jpg" alt="Snow" color="black" style="width:500px" border="2px">



<span style="font-style:italic; color:#ff8c00">***We also use another CDN before CloudCDN.***</span>


## 5. Verification 

Test your each URL https://YOUR-DOMAIN/

You can see your client IP will show on the browser as well.

<img src="https://storage.googleapis.com/stanleymdbucket/md_image/detect-IP-GKE-CDN/detecet-IP-GKE-CDN-08.jpg" alt="Snow" color="black" style="width:1000px" border="2px">


And we can check the IP is correct.

<img src="https://storage.googleapis.com/stanleymdbucket/md_image/detect-IP-GKE-CDN/detecet-IP-GKE-CDN-10.jpg" alt="Snow" color="black" style="width:1000px" border="2px">
 
