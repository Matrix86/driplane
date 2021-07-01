---
title: "Slack"
date: 2020-10-12T19:38:02+02:00
draft: false
---

## Slack

This filter allows you to send files, messages and download files from Slack. 
It can be used alone or with the Slack Feeder to create a simple Slack bot for events API.  

### Parameters

The following parameters are required from this filter:

 | Parameter | Type | Default | Description 
 | --- | --- | --- | --- |
 | **action** | _STRING_ | "send_message" | action to perform: "send_message", "send_file", "download_file", "user_info" |
 | **token** | _STRING_ | "" | Slack bot Token |
 
Each `action` can use different parameters:

#### action = send_message

 | Parameter | Type | Default | Description 
 | --- | --- | --- | --- |
 | **to** | _STRING_ | "" | channel ID or User ID that should receive the message (supports [Golang templates](https://golang.org/pkg/text/template/)) |
 | **target** | _STRING_ | "main" | if `text` is not used you can choose which field of `Message` to use as text |
 | **text** | _STRING_ | "" | you can define the text of the message to send (supports [Golang templates](https://golang.org/pkg/text/template/)) | 
 | **blocks** | _BOOL_ | false | if true you can use the Slack template blocks ([block builder](https://api.slack.com/tools/block-kit-builder)) | 

#### action = send_file

 | Parameter | Type | Default | Description 
 | --- | --- | --- | --- |
 | **to** | _STRING_ | "" | channel ID or User ID that should receive the message (supports [Golang templates](https://golang.org/pkg/text/template/)) |
 | **target** | _STRING_ | "main" | if `filename` is not specified you can choose which field of `Message` has to be used as file content to send |
 | **filename** | _STRING_ | "" | path of the file to send (supports [Golang templates](https://golang.org/pkg/text/template/)) |

#### action = download_file

 | Parameter | Type | Default | Description 
 | --- | --- | --- | --- |
 | **url** | _STRING_ | "" | it should contain the Slack private url for the file download (supports [Golang templates](https://golang.org/pkg/text/template/)) |
 | **target** | _STRING_ | "urlprivate" | if `url` is not specified you can choose which field of `Message` contains the Slack private url |
 | **filename** | _STRING_ | "" | path of where to save the downloaded file. If not specified the file content will be inserted in the `main` field (supports [Golang templates](https://golang.org/pkg/text/template/)) |
 
#### action = user_info

 | Parameter | Type | Default | Description 
 | --- | --- | --- | --- |
 | **target** | _STRING_ | "user" | this field has to contain the USERID |
 
In the output Message you can find all the user information returned by Slack:

```
	user_id                
	user_teamid            
	user_name              
	user_deleted           
	user_color             
	user_realname          
	user_tz                
	user_tzlabel           
	user_tzoffset          
	user_profile           
	user_isbot             
	user_isadmin           
	user_isowner           
	user_isprimaryowner    
	user_isrestricted      
	user_isultrarestricted 
	user_isstranger        
	user_isappuser         
	user_isinviteduser     
	user_has2fa            
	user_hasfiles          
	user_presence          
	user_locale            
	user_updated           
	user_enterprise        
```

{{< notice info "Example" >}} 
`... | slack(action="user_info", target="user") | slack(action="send_message", to="{{.channel}}", text="Hi {{.user_realname}}") | ...`
`... | random(output="random_num", min="9", max="90") | slack(action="download_file", filename="/tmp/store_{{ .random_num }}.data") | ...`
{{< /notice >}}

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 