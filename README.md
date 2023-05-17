<p align="center">
  <img src="https://github.com/Matrix86/driplane/blob/gh-pages/logo.png"/>
</p>


<p align="center">
  <a href="https://www.buymeacoffee.com/mtx86">
    <img src="https://img.shields.io/badge/Buy me a-%F0%9F%8D%BA%20beer-blue?logo=beer&style=for-the-badge&color=grey&logoColor=white&labelColor=blue">
  </a>
</p>

# Driplane ![License](https://img.shields.io/github/license/Matrix86/driplane) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/Matrix86/driplane) ![GitHub Workflow Status (branch)](https://img.shields.io/github/workflow/status/Matrix86/driplane/Build%20and%20Test/master) ![GitHub release (latest by date)](https://img.shields.io/github/v/release/Matrix86/driplane?color=red) ![Codecov](https://img.shields.io/codecov/c/github/Matrix86/driplane) 

`Driplane` allows you to create an automatic alerting system or start automated tasks triggered by events.
You can keep under control a stream source as Twitter, a file, a RSS feed or a website.
It includes a mini language that allows you to define the filtering pipelines (the rules), and it is extensible thanks to an integrated JS plugin system. 

With `driplane` you can create several rules starting from one or more streams, filter the content in the pipeline and launch tasks or send alerts if some event occurred.

The complete documentation can be found [HERE](https://matrix86.github.io/driplane/doc/)

## Examples

### Twitter

Keep under control the Twitter account of some users or keywords used in Tweets, search for hashes and send a message on Slack using the webhook. 
The cache will avoid sending messages with hashes already seen in the last 24h. 

<details>
<summary><b>Show example</b></summary>

**`twitter.rule`**
```bash
# Twitter feed
# Define a rule with a Twitter feeder and define keywords and users
Twitter => <twitter: users="goofy, mickeymouse", keywords="malware, virus, PE">;

# Define a rule to send a slack message using a defined api hook
slack => http(url="https://hooks.slack.com/services/XXXXXXXXXX/XXXXXXXXXX/XXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",method="POST",headers="{\"Content-type\": \"application/json\"}",rawData="{{.main}}");

# Define a rule that filter the received tweets
tweet_rule => @Twitter |
              # ignore spanish tweets
              !text(target="language", pattern="es") |
              # extract hashes from them
              hash(extract="true") |
              # add a new field to the stream with the hash
              override(name="hash", value="{{ .main }}") |
              # drop it if we saw that hash before
              cache(ttl="24h", global="true") |
              # fill the template with extracted data
              format(file="slack_twitter.txt") |
              # use the rule defined above to send the filled template to slack endpoint
              @slack_alert;
```

**`slack_twitter.txt`**
```json
{
	"blocks": [{
		"type": "context",
		"elements": [{
			"text": "*Rule* : {{.rule_name}} | *Feeder* : {{.source_feeder}} ",
			"type": "mrkdwn"
		}]
	}, {
		"type": "divider"
	}, {
		"type": "section",
		"text": {
			"type": "mrkdwn",
			"text": "Found a new hash : _{{.hash}}_\nLink to the Twitter post: {{ .link }}"
		}
	}]
}
```

</details>

### RSS Feed

Check a RSS feed with a defined frequency to be alerted every time a news containing one or more interesting keywords are published, and send a telegram message.

<details>
<summary><b>Show example</b></summary>

**`rss.rule`**
```bash
# Feed example
# Define a rule called 'RSS' that read a RSS feed every minutes
RSS => <rss: url="http://rss.cnn.com/rss/cnn_topstories.rss", freq="1m", ignore_pubdate="true">;

# Define a rule to send a telegram message using a defined api hook
telegram =>  http(url="https://api.telegram.org/XXX:XXXX/sendMessage", method="POST", headers="{\"Content-type\": \"application/json\"}", rawData="{{.main}}");

news => @RSS |
        # skip links if we saw that before
        cache(ttl="100h", target="link") |
        # Search in the description field using a regular expression
        text(pattern="(?i)tech|discovery|bitcoin|trump", regexp="true", target="description") |
        # format the output text to send on telegram
        format(template="Found new interesting article: {{ .link }}") |
        @telegram;
```

</details>

### Slack

Creates a simple bot or keep under control one or more channel. Using the event APIs of Slack every time the bot receives 
a message with hashes, it will try to get information from this hash and it replies to the original channel 
(in a private chat if the user contacted the bot privately or in a channel if it has been added to a channel and the event comes from there) with all the gathered info.
Also, using another rule, if a file is uploaded, the bot will analyze it and return a report in the reply message. 

<details>
<summary><b>Show example</b></summary>

**`slack.rule`**
```bash
# Simple Slack Bot
# define the slack feeder: token and verification token are defined in the configuration file
SlackEvent => <slack>;

# Get status from zMD
status => @SlackEvent |
          # consider only message events
          text(target="type", pattern="message") |
          # extract all the hashes found in the message
          hash(extract="true") |
          # logic to get the info in the report
          js(path="bot.js", function="GetHashReport") |
          # format of the response using the Slack template system
          format(file="slack_report.txt") |
          # reply to the channel where the event has been generated 
          slack(action="send_message", to="{{.channel}}", target="main", blocks="true");

# Upload file and get status
upload => @SlackEvent |
          # consider only file_share events
          text(target="type", pattern="file_share") |
          # download the file and store it in /tmp/nameofthefile
          slack(action="download_file", filename="/tmp/{{ .name }}") |
          # call the method UploadFile() in bot.js: it extract info from the file and return them
          js(path="bot.js", function="UploadFile") |
          # format of the response using the Slack template system
          format(file="slack_report.txt") |
          # reply to the channel where the event has been generated 
          slack(action="send_message", to="{{.channel}}", target="main");

```

</details>

## How it works

The user can define one or more rules. Usually a rule contains a source (`feeder`), who cares of getting the information and sending updates (`Message`) through the pipeline, and one or more `filters`.
The filters' job is to choose whether to propagate or not the `Message` to the next filter in the pipeline relying on a _condition_, or change the `Message` received before to propagate it. The `Message` will be propagated only if it verifies the condition.

## Use cases

Using `driplane` is it possible to:

 * keep track of keywords or users on Twitter, receive the new tweets or quoted tweets from them, search for URLs or particular strings in them and send a Telegram or a Slack message through their webhooks.
 * keep track of a RSS feed or a website, and download and store on file all the new changes to them.
 * keep track of changes on a file, and launch alert if a particular condition is verified.
 
The rules and the JS plugins allows you to create very complex custom logics.
  
## Usage

```
Usage of ./bin/driplane:
  -config string
    	Set configuration file.
  -debug
    	Enable debug logs.
  -dry-run
    	Only test the rules syntax.
  -help
    	This help.
  -js string
    	Path of the js plugins.
  -rules string
    	Path of the rules' directory.
```

