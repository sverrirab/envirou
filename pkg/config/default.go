package config

import (
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

const default_ini = `
; Default configuration file for envirou - feel free to edit!
; (If you remove it a new one will be generated).

; ── Settings ─────────────────────────────────────────────────

[settings]
quiet=0
sort_keys=1
path_tilde=1  ; display only: replaces $HOME with ~ in output
password=AWS_SECRET_ACCESS_KEY, AWS_SESSION_TOKEN
path=HOME, PATH, GOPATH, JAVA_HOME, KUBECONFIG, VIRTUAL_ENV

; ── Display colors ───────────────────────────────────────────
; Valid colors: green, magenta, red, yellow, blue, cyan, white,
;               black, bold, underline, reverse, deleted, none

[format]
group=magenta
profile=green
env_name=cyan
path=reverse

; ── Visible groups ───────────────────────────────────────────
; Use * for wildcards. These groups are shown by default.

[groups]
basic=PATH
aws=AWS_*, EC2_*
cloud=KUBECONFIG, KUBE_*, DOCKER_*, COMPOSE_*, TF_*, TERRAFORM_*
dev=GOROOT, GOPATH, GOBIN, JAVA_HOME, JAVA_OPTS, JDK_HOME, CLASSPATH, MAVEN_HOME, GRADLE_HOME, VIRTUAL_ENV, PYTHONPATH, PYTHONHOME, CONDA_*, PYENV_*, NODE_*, NPM_*, NVM_*, CARGO_HOME, RUSTUP_HOME, GEM_HOME, RBENV_*, BUNDLE_*
git=GIT_*

; ── Hidden groups (. prefix) ─────────────────────────────────
; Shown only with -a flag. Split these in [custom] if you need finer control.

.shell=BASH_*, BASH, BASHPID, BASHOPTS, BASH_ENV, COMP_WORDBREAKS, DIRSTACK, EPOCHREALTIME, EPOCHSECONDS, FUNCNAME, GROUPS, HISTCMD, LINENO, MACHTYPE, OPTARG, OPTIND, OSTYPE, PIPESTATUS, SHELLOPTS, SHLVL, ZSH_*, ZSH, ZSH_NAME, ZSH_VERSION, ZDOTDIR, ZLE_*, RPROMPT, RPS1, PROMPT, PROMPT2, PROMPT3, PROMPT4, PROMPT_EOL_MARK, PSVAR, PS*, SECONDS, RANDOM, _, COLUMNS, LINES, TTY, HIST*, SAVEHIST, MAIL, MAILCHECK, UID, EUID, PPID
.locale=LANG, LANGUAGE, LC_*, LINGUAS
.network=SSH_*, TMUX, TMUX_*, STY, WINDOW
.system=XDG_*, SUDO_*, COMMAND_MODE, DISPLAY, MOTD_SHOWN, PULSE_SERVER, WAYLAND_DISPLAY, SECURITYSESSIONID, LOGNAME, NAME, USER, TMP, TMPDIR, HOME, EDITOR, SHELL, INFOPATH, HOMEBREW_*
.terminal=TERM, TERM_PROGRAM, TERM_PROGRAM_VERSION, TERMCAP, TERMINFO_*, TERM_FEATURES, COLORTERM, COLORFGBG, LSCOLORS, LS_COLORS, LESS, LESSCLOSE, LESSOPEN, PAGER
.editors=VSCODE_*, VSCODE_GIT_ASKPASS_*, GIT_ASKPASS
.macos=__CF_*, XPC_*, TERM_SESSION_ID, LC_TERMINAL, LC_TERMINAL_VERSION, OSLogRateLimit, SQLITE_EXEMPT_PATH_FROM_VNODE_GUARDS, ITERM_*, VTE_*, Apple_PubSub_Socket_Render, LaunchInstanceID
.windows=TEMP, USERNAME, USERPROFILE, USERDOMAIN*, OS, LOGONSERVER, COMPUTERNAME, HOMEDRIVE, HOMEPATH, PUBLIC, APPDATA, LOCALAPPDATA, PROGRAMDATA, PROGRAMFILES, PROGRAMFILES(X86), PROGRAMW6432, COMMONPROGRAMFILES, COMMONPROGRAMFILES(X86), COMMONPROGRAMW6432, DRIVERDATA, SYSTEMDRIVE, SYSTEMROOT, WINDIR, NUMBER_OF_PROCESSORS, PROCESSOR_*, ALLUSERSPROFILE, PSMODULEPATH, FP_NO_HOST_CHECK, PATHEXT, OneDrive*, COMSPEC, CMDCMDLINE, CMDEXTVERSION, ERRORLEVEL, SESSIONNAME, CLIENTNAME, WT_*, HOSTTYPE, WSLENV, WSL_DISTRO_NAME, WSL_*, WSL2_*
.powershell=PWD, PID, HOME, Host, Error, Args, Input, Matches, MyInvocation, NestedPromptLevel, LASTEXITCODE, ShellId, StackTrace, PSItem, this, foreach, switch, Event, EventArgs, EventSubscriber, Sender, PSSenderInfo, IsWindows, IsLinux, IsMacOS, IsCoreCLR, EnabledExperimentalFeatures, true, false, null

; ── Ignored groups (.. prefix) ───────────────────────────────
; Hidden and excluded from snapshot/diff.

..ignore=_, PWD, OLDPWD, SHLVL

; ── Custom ───────────────────────────────────────────────────
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

const configFileName = "config.ini"
const snapshotFileName = "snapshot.ini"

// GetDefaultConfigFilePath Returns full path to the config file
func GetDefaultConfigFilePath() string {
	full_path := filepath.Join(GetDefaultConfigFileFolder(), configFileName)
	return full_path
}

// GetSnapshotFilePath returns the full path to the snapshot file
func GetSnapshotFilePath() string {
	return filepath.Join(GetDefaultConfigFileFolder(), snapshotFileName)
}

// GetDefaultConfigFileFolder Figures out where the config file should be
func GetDefaultConfigFileFolder() string {
	current_user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Join(current_user.HomeDir, ".config", "envirou")
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
