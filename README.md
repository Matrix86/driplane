![Logo](https://github.com/Matrix86/driplane/blob/master/docs/logo.png)
![License](https://img.shields.io/github/license/Matrix86/driplane)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/Matrix86/driplane)
![GitHub Workflow Status (branch)](https://img.shields.io/github/workflow/status/Matrix86/driplane/Build%20and%20Test/master)

`Driplane` allows you to create an automatic alerting system or start automated tasks triggered by events.
You can keep under control a stream source as Twitter, a file, a RSS feed or a website.
It includes a mini language that allows you to define the filtering pipelines (the rules), and it is extensible thanks to an integrated JS plugin system. 

With `driplane` you can create several rules starting from one or more streams, filter the content in the pipeline and launch tasks or send alerts if some event occurred.

The documentation can be found [HERE](https://matrix86.github.io/driplane/doc/)

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
Usage of driplane:
  -config string
    	Set configuration file.
  -debug
    	Enable debug logs.
  -help
    	This help.
  -js string
    	Path of the js plugins.
  -rules string
    	Path of the rules' directory.
```