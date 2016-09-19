# f1le

**f1le** is a simple site for hosting files. It is designed to be used by a single person or a small group of people.

# Usage

These instructions apply to you only if you wish to run your own instance of f1le.

First, you must have Go installed and configured. Next, you can download and install the f1le server and its dependencies as follows:

    $ go get github.com/gorilla/securecookie
	$ go get github.com/gorilla/sessions
    $ go get github.com/hoisie/mustache
    $ go get github.com/unixpickle/f1le
    $ go install github.com/unixpickle/f1le

In order to run the server, you will have to create a directory in which files will be stored. Once you've done that, you can run f1le. The first time you run f1le, it will ask you to set a password.

    $ mkdir cool_files
    $ f1le 1337 cool_files

Now you have f1le up and running. Congrats!

## Docker

You can build/run this with Docker. If you have Docker installed and running  
simply download the Dockerfile to a directory, cd there, and build the image:

    docker build -t f1le .

Then run it:

    docker run -it -p 8080:8080 f1le

You will be prompted for a password.

# TODO

Here are some things which need to get done before this is finished:

 * Click to choose file
 * Better alert dialogs (don't use `window.alert` and `window.confirm`)
 * Show the amount of free space on the destination volume if possible
 * Temporary upload access
