# envirou - View and manage your shell environment

![Build Status](https://travis-ci.org/sverrirab/envirou.svg?branch=master)

Envirou (`ev`) helps you to quickly view and configure your shell 
 environment. Display important variables with nice formatting and hide the ones you don't care about. No more custom shell scripts to configure your environment or guessing which one is active!
 

![Simple View](./screenshots/header.png)


# Key hightlights 
* Works with any other tool - just views and optionally sets environment variables.
* Compact output (replaces $HOME with `~` and _underscores_ paths for readability).
* Hides all irrelevant variables such as `TMPDIR`, `LSCOLORS` etc, etc.
* Fully customizable.
* Works on Mac + Linux (bash + zsh) and Windows.  
* Fully standalone with no dependencies except any python 2.7 or 3.4+ you have installed.
* Command completion support (bash + zsh).
* Includes [oh-my-zsh](https://ohmyz.sh/) theme.


## Why?
Everyone that works with complex infrastructure or multiple development environments from the command line know the feeling of using the wrong toolchain or environment and having the nagging suspicion that you have mixed something up in your configuration. Classical examples 
are PATH's to tools/SDK versions, external service endpoints for your PROD and DEV environments
etc etc.

Most tools have some way of switching between environments but that has two problems - you have to learn the idiosyncrasies of each one and often it is very opaque what configuration is currently in effect.
Most tools are configurable using environment variables and Envirou allows you to quickly switch and display what configuration is currently effect.


## Quickstart

### Using Mac OS X or Linux

```bash
$ curl -o- https://raw.githubusercontent.com/sverrirab/envirou/master/curl_install.sh | bash
```

### Using Windows

1) Make sure you have python installed (`py` should work, 3.7+ recommended).
1) Download the [zip file](https://github.com/sverrirab/envirou/archive/master.zip) (or do git checkout).
3) Extract the zip file and open a shell in that folder.

```cmd
> echo @%CD%\envirou.bat %* > <tools-folder-on-path>\ev.bat 
```

## Example use cases
### AWS configuration
Add all your aws profiles in `~/.aws/config`.  Then create an Envirou profile for each
that sets the `AWS_PROFILE` variable to the name of the AWS profile.  This way you can
easily e.g. switch between `dev` and `prod` or even the default region or output formatting.

### Kubectl configuration
Copy your `~/.kube/config` into a new file for each environment.  Create an Envirou 
profile for each that sets the `KUBECONFIG` pointing to each file.

Make sure you set the default context in each file to be the correct one.  This way you
can create different profiles that for example have a different default namespace.

## Advanced configuration

Edit the ini file:

```bash
$ ev --edit     # Customize settings.
```

Example from `envirou.ini`:

```inifile
[profile:basic]
PATH=/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:/Users/sab/src/custom/bin
VIRTUAL_ENV
AWS_PROFILE

[profile:py3]
PATH=/Users/sab/sdk/py3/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:/Users/sab/src/custom/bin
VIRTUAL_ENV=/Users/sab/sdk/py3

[profile:py2]
PATH=/Users/sab/sdk/py2/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:/Users/sab/src/custom/bin
VIRTUAL_ENV=/Users/sab/sdk/py2

[profile:awsprod]
AWS_PROFILE=prod

[profile:awsdev]
AWS_PROFILE=dev
```


## Where does the name come from? 
The name Envirou is inspired by Spirou the comic book character.  
The alias `ev` is both short for *Envirou* and `env`. 


## License

Free for any use see [MIT License](./LICENSE) for details.
