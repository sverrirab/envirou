package config

import (
	"io/ioutil"
	"os"
	"path"
)

const default_ini = `
; Default configuration file for envirou - feel free to edit!
; (If you remove it a new one will be generated).

[settings]
quiet=0
sort_keys=1
path_tilde=1
password=AWS_SECRET_ACCESS_KEY, AWS_SESSION_TOKEN
path=HOME, PATH, GOPATH, JAVA_HOME, KUBECONFIG, VIRTUAL_ENV

[format]
; <color> can be one of: green, magenta, red, yellow, blue, bold, underline, none
group=magenta
profile=green
env_name=cyan

[groups]
path=PATH, PWD, TMP, TMPDIR, HOME, EDITOR, GOROOT, GOPATH, JAVA_HOME, VIRTUAL_ENV

; Use * at the end to match multiple:
aws=AWS_*, EC2_*
golang=GOROOT, GOPATH

; Names starting with . are hidden by default.
.basic=LOGNAME, NAME, USER, TMP, TMPDIR, HOME, EDITOR, MAIL
.term=ZSH, SHELL, COLORTERM, LSCOLORS, LS_COLORS, LESS, LESSCLOSE, LESSOPEN, PAGER, TERM, TERM_PROGRAM, TERM_PROGRAM_VERSION, LC_*, LANG, LANGUAGE
.system=SSH_*, XDG_*, XPC_*, SUDO_*, COMMAND_MODE, SECURITYSESSIONID

; Names starting with .. are hidden and not used in diff/reset to default.
..ignore=_, PWD, OLDPWD, SHLVL, PS1, PROMPT, SESSIONNAME, TERM_SESSION_ID, ITERM_*, COLORFGBG, COLORTERM

; Apple specific 
.apple=Apple_PubSub_Socket_Render, __CF_USER_TEXT_ENCODING, LaunchInstanceID, __CFBundleIdentifier

; Windows specific below
.winbasic=TEMP, USERNAME, USERPROFILE, USERDOMAIN*, LOGONSERVER, COMPUTERNAME, HOMEDRIVE, HOMEPATH, PUBLIC, APPDATA, LOCALAPPDATA 
.windows=HOSTTYPE, WSLENV, WSL_DISTRO_NAME, MOTD_SHOWNOS, COMSPEC, PROGRAMDATA, PROGRAMFILES, PROGRAMFILES(X86), PROGRAMW6432, COMMONPROGRAMFILES, COMMONPROGRAMFILES(X86), COMMONPROGRAMW6432, DRIVERDATA, SYSTEMDRIVE, SYSTEMROOT, WINDIR, NUMBER_OF_PROCESSORS, PROCESSOR_*, ALLUSERSPROFILE, PSMODULEPATH, FP_NO_HOST_CHECK, PATHEXT

; Add your customizations below this point.

[custom]
; Add custom groups here.
; example=EXAMPLE_*

; Add your own profiles here...
; [profile:example]
; EXAMPLE_OCCUPATION=elevator operator
; EXAMPLE_EMPTY_VARIABLE=
; EXAMPLE_UNSET_VARIABLE

`

const configFileNmae = "config-v2.ini"

// GetDefaultConfigFilePath Returns full path to the config file
func GetDefaultConfigFilePath() string {
	full_path := path.Join(GetDefaultConfigFileFolder(), configFileNmae)
	return full_path
}

// GetDefaultConfigFileFolder Figures out where the config file should be
func GetDefaultConfigFileFolder() string {
	home := os.Getenv("HOME")
	return path.Join(home, ".config", "envirou")
}

// WriteDefaultConfigFile write the default config file if no file exists already
func WriteDefaultConfigFile(path string) error {
	// Make sure the folder exists
	err := os.MkdirAll(GetDefaultConfigFileFolder(), os.ModePerm)
	if err != nil {
		return err
	}
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		err = ioutil.WriteFile(path, []byte(default_ini), 0644)
	}
	return err
}
