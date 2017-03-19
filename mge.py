#!/usr/bin/env python
from __future__ import print_function
import argparse
import os
import sys
from collections import defaultdict

_CONSOLE_COLORS = {
    "c-end":        "\033[0m",
    "c-bold":       "\033[1m",
    "c-underline":  "\033[4m",
    "c-fail":       "\033[91m",
    "c-green":      "\033[92m",
    "c-warn":       "\033[93m",
    "c-blue":       "\033[94m",
    "c-header":     "\033[95m",
}

_DEFAULT_CONFIG = """; Default configuration file for mge - feel free to edit.
[groups]
; Names starting with _ are hidden by default
AWS=AWS_ACCESS_KEY,AWS_SECRET_KEY,EC2_HOME,EC2_URL,S3_URL
CloudStack=CS_API_URL,CS_API_KEY,CS_SECRET_KEY,CS_URL
GO=GOPATH
Java=JAVA_HOME
Python=VIRTUAL_ENV
Standard=PATH
_Apple=Apple_PubSub_Socket_Render,__CF_USER_TEXT_ENCODING,XPC_FLAGS,XPC_SERVICE_NAME
_Python=VERSIONER_PYTHON_PREFER_32_BIT,VERSIONER_PYTHON_VERSION
_mge=MGE_HOME
_iTerm=ITERM_PROFILE,ITERM_SESSION_ID,COLORFGBG
_SSH=SSH_AUTH_SOCK
_Env=TERM,SHELL,LOGNAME,USER,HOME,EDITOR,LC_ALL,LC_CTYPE,TMP,TMPDIR
_Terminal=TERM_PROGRAM,TERM_PROGRAM_VERSION,TERM_SESSION_ID
_Shell=_,PWD,SHLVL
_Progs=LSCOLORS,PAGER,LESS,ZSH
[profile:stuff1]
PROF_NAME=stuff1
PROF_XYZ=abc
[profile:stuff2]
PROF_NAME=stuff2
PROF_XYZ=xyz
"""

_CONFIG_PATH = "~/.mge"
_CONFIG_FILE = "config"
_SECTION_GROUPS = "groups"
_SECTION_PROFILE = "profile"
_NOGROUP = "NA"

_verbose_level = 0
_groups = defaultdict(list)
_profiles = defaultdict(list)
_stdout = None


def redirect_stdout():
    global _stdout
    _stdout = sys.stdout
    sys.stdout = sys.stderr


def very_verbose(fmt, *args, **kwargs):
    if _verbose_level > 0:
        verbose(fmt, *args, **kwargs)


def verbose(fmt, *args, **kwargs):
    kwargs["file"] = sys.stderr
    kwargs.update(**_CONSOLE_COLORS)
    output = fmt.format(**kwargs)
    print(output, *args)


def output_group(group):
    if group == _NOGROUP:
        group += " (No Applicable group)"
    verbose("{c-header}# {group} {c-end}", group=group)


def output_key(k, maxlen):
    verbose("{key:<{maxlen}} = {c-bold}{value}{c-end}", key=k, value=os.environ[k], maxlen=maxlen)


def read_config():
    folder = os.environ.get("MGE_HOME", "")
    if len(folder.strip()) == 0:
        folder = os.path.expanduser(_CONFIG_PATH)
    if not os.path.isdir(folder):
        very_verbose("Creating configuration folder:", folder)
        os.makedirs(folder)
    config = os.path.join(folder, _CONFIG_FILE)
    if not os.path.exists(config):
        very_verbose("First time initializization of config file:", config)
        with open(config, "w") as f:
            f.write(_DEFAULT_CONFIG)

    with open(config, "r") as f:
        section = "(none)"
        for l in f.readlines():
            l = l.split(";")[0].split("#")[0].strip()
            if len(l) == 0:
                continue
            if l[0] == "[" and l[-1] == "]":
                section = l[1:-1]
                continue
            key, value = l.split("=", 1)

            if section == _SECTION_GROUPS:
                for env in value.split(","):
                    very_verbose(_SECTION_GROUPS, key, env)
                    _groups[key].append(env.strip())
            elif section.startswith(_SECTION_PROFILE):
                _, profile = section.split(":", 1)
                very_verbose(_SECTION_PROFILE, profile, key, value)
                _profiles[profile].append(l)
            else:
                very_verbose("Ignoring config item:", section, key, value)


def main(arguments):
    read_config()
    match_group = defaultdict(list)
    for name, keys in _groups.items():
        for k in sorted(keys):
            match_group[k].append(name)

    maxlen = max([len(k) for k in os.environ.keys()])
    grouped = defaultdict(list)
    for k in sorted(os.environ.keys()):
        matched_groups = match_group[k]

        if matched_groups:
            for group in matched_groups:
                grouped[group].append(k)
        else:
            grouped[_NOGROUP].append(k)

    filter_groups = len(arguments.group) > 0
    not_displayed_group = []
    for group in sorted(grouped.keys()):
        is_hidden = (group[0] == "_")
        if arguments.all or (filter_groups and group in arguments.group) or (not filter_groups and not is_hidden):
            output_group(group)
            for k in grouped[group]:
                output_key(k, maxlen)
        else:
            not_displayed_group.append(group)

    if len(not_displayed_group):
        output_group("mge [-h|-a] {}".format(" ".join(not_displayed_group)))

    if arguments.profile:
        for l in _profiles[arguments.profile]:
            key, value = l.split("=", 1)
            verbose(key, value)

    # TODO: REMOVE
    # print("export MUCHADOABOUTNOTHING=1", file=_stdout)

    return 0


if __name__ == "__main__":
    redirect_stdout()
    parser = argparse.ArgumentParser()
    parser.add_argument("-v", "--verbose", action="count", default=0,
                        help="Increase output verbosity")

    parser.add_argument("-a", "--all", dest="all", action="store_true", help="Show hidden groups")
    parser.add_argument("--no-all", dest="password", action="store_false", help="Don't show hidden groups (default)")
    parser.set_defaults(all=False)

    parser.add_argument("-p", "--profile", help="Activate profile")
    parser.add_argument("group", nargs="*", help="Display named group(s)")

    args = parser.parse_args()
    _verbose_level = args.verbose
    exit(main(args))
