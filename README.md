<p align="center">
  <img alt="Logo" src="https://github.com/Matrix86/driplane/blob/master/docs/logo.png"/>
  <p align="center">
  Driplane
  </p>
</p>

Driplane allows you to create an automatic alerting system or start automated tasks triggered by events.
You can keep under control a source as Twitter, a file, a RSS feed or a website.
It includes a mini language to define the filtering pipelines (the rules) and contains a JS plugin system. 

With driplane you can create several rules starting from one or more streams, filter the content in the pipeline and launch task or send e-mail if some events occurred.

The documentation can be found [HERE](https://matrix86.github.io/driplane/)

## How it works

The user can define one or more rules. Each rules contain a source (`feeder`), who cares of get the information and send updates (`Message`) through the pipeline, and several `filters`.
The task of the filters is to propagate or stop the `Message` to the next filter in the pipeline, or change the `Message` received before to propagate it. If the condition is verified, the `Message` will be propagated.

## Use cases

Using driplane is it possible to:

 * keep track of keywords or users on Twitter, receive the new tweets or quoted tweet from them, search for URLs or particular strings and send it to a Telegram or Slack channel through a webhook.
 * keep track of a RSS feed or a website, and download and store on file all the new changes to them.
 
The rules and the JS plugins allows you to create more complex custom logics.
  
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