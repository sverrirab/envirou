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
    "c-red":        "\033[31m",
    "c-green":      "\033[32m",
    "c-yellow":     "\033[33m",
    "c-blue":       "\033[34m",
    "c-magenta":    "\033[35m",
}

_CONFIG_ENV = "ENVIROU_HOME"
_CONFIG_PATH = "~/.envirou"
_CONFIG_FILE = "config"
_SECTION_GROUPS = "groups"
_SECTION_CUSTOM = "custom"
_SECTION_HIGHLIGHT = "highlight"
_SECTION_PROFILE = "profile"
_NOGROUP = "NA"

_verbose_level = 0
_groups = defaultdict(list)
_highlight = {}
_profiles = defaultdict(list)
_stdout = None


def redirect_stdout():
    global _stdout
    _stdout = sys.stdout
    sys.stdout = sys.stderr


def shell_eval(fmt, *args, **kwargs):
    if _verbose_level > 0:
        verbose("SHELL> " + fmt, *args, **kwargs)
    output = fmt.format(**kwargs)
    print(output, *args, file=_stdout)


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
    verbose("{c-magenta}# {group} {c-end}", group=group)


def output_key(k, maxlen):
    fmt = "{key:<{maxlen}} {value}"
    if k in _highlight:
        fmt = "{c-" + _highlight.get(k) + "}" + fmt + "{c-end}"
    verbose(fmt, key=k, value=os.environ[k], maxlen=maxlen)


def read_config():
    folder = os.environ.get("ENVIROU_HOME", "")
    if len(folder.strip()) == 0:
        folder = os.path.expanduser(_CONFIG_PATH)
    if not os.path.isdir(folder):
        very_verbose("Creating configuration folder:", folder)
        os.makedirs(folder)
    config = os.path.join(folder, _CONFIG_FILE)
    if not os.path.exists(config):
        very_verbose("First time initializization of config file:", config)
        py_path = os.path.realpath(__file__)
        config_path = py_path[:-3] + ".default"
        very_verbose("Reading from template:", py_path)
        with open(config_path, "r") as template:
            default_config = template.read()

        with open(config, "w") as f:
            f.write(default_config)

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

            if section == _SECTION_GROUPS or section == _SECTION_CUSTOM:
                for env in value.split(","):
                    very_verbose(_SECTION_GROUPS, key, env)
                    _groups[key].append(env.strip())
            elif section == _SECTION_HIGHLIGHT:
                for env in value.split(","):
                    _highlight[env] = key
            elif section.startswith(_SECTION_PROFILE):
                _, profile = section.split(":", 1)
                very_verbose(_SECTION_PROFILE, profile, key, value)
                _profiles[profile].append(l)
            else:
                very_verbose("Ignoring config item:", section, key, value)


def main(arguments):
    read_config()

    if arguments.reset:
        very_verbose("resetting standard profile")

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
        is_hidden = (group[0] == ".")
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
    shell_eval("export MUCHADOABOUTNOTHING=1")

    return 0


if __name__ == "__main__":
    redirect_stdout()
    parser = argparse.ArgumentParser()
    parser.add_argument("-v", "--verbose", action="count", default=0,
                        help="Increase output verbosity")

    parser.add_argument("-a", "--all", dest="all", action="store_true", help="Show hidden groups")
    parser.add_argument("--no-all", dest="password", action="store_false", help="Don't show hidden groups (default)")
    parser.set_defaults(all=False)

    parser.add_argument("-r", "--reset", help="Activate profile")
    parser.add_argument("-p", "--profile", help="Activate profile")
    parser.add_argument("group", nargs="*", help="Display named group(s)")

    args = parser.parse_args()
    _verbose_level = args.verbose
    exit(main(args))
