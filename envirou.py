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
_DEFAULT_FILE = "default.envirou"
_SECTION_GROUPS = "groups"
_SECTION_PROFILES = "profiles"
_SECTION_CUSTOM = "custom"
_SECTION_HIGHLIGHT = "highlight"
_SECTION_PROFILE_START = "profile:"
_HIGHLIGHT_PASSWORD = "password"
_NOGROUP = "NA"

_verbose_level = 0
_groups = defaultdict(list)
_profiles = defaultdict(dict)
_highlight = {}
_default = {}
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


def ultra_verbose(fmt, *args, **kwargs):
    if _verbose_level > 1:
        verbose(fmt, *args, **kwargs)


def very_verbose(fmt, *args, **kwargs):
    if _verbose_level > 0:
        verbose(fmt, *args, **kwargs)


def verbose(fmt, *args, **kwargs):
    kwargs["file"] = sys.stderr
    kwargs.update(**_CONSOLE_COLORS)
    output = fmt.format(**kwargs)
    print(output, *args)


def color_wrap(s, color):
    return "{c-" + color + "}" + s + "{c-end}"


def output_group(group):
    if group == _NOGROUP:
        group += " (No Applicable group)"
    out = color_wrap("# {group}", color=_highlight.get(_SECTION_GROUPS, "magenta"))
    verbose(out, group=group)


def output_key(k, maxlen, password=False):
    fmt = "{key:<{maxlen}} {value}"
    value = os.environ.get(k, "")
    if _default and ((k in _default and value != _default[k]) or k not in _default):
        fmt = "{c-red}" + fmt + "{c-end}"
    elif k in _highlight:
        color = _highlight.get(k)
        if color == _HIGHLIGHT_PASSWORD:
            if not password:
                value = "*" * len(value)
        else:
            fmt = color_wrap(fmt, color)
    verbose(fmt, key=k, value=value, maxlen=maxlen)


def output_profiles(active, inactive):
    def_color = _highlight.get(_SECTION_GROUPS, "magenta")
    active_color = _highlight.get(_SECTION_PROFILES, "yellow")
    active_str = color_wrap(", ", def_color).join([color_wrap(p, active_color) for p in active])
    inactive_str = ", ".join(inactive)
    s = ""
    if active:
        s = color_wrap("# Active profiles: ", def_color) + active_str

    if inactive and active:
        s += color_wrap(" (inactive: {})".format(inactive_str), def_color)
    elif inactive:
        s = color_wrap("# Inactive profiles: {}".format(inactive_str), def_color)

    if s:
        verbose(s)


def clean_split(s, sep="="):
    k, v = s.split(sep, 1)
    return k.strip(), v.strip()


def config_filename(short):
    """
    :param short: short name
    :return: Get full path to config filename.
    """
    folder = os.environ.get(_CONFIG_ENV, _CONFIG_PATH)
    folder = os.path.expanduser(folder)
    if not os.path.isdir(folder):
        very_verbose("Creating configuration folder:", folder)
        os.makedirs(folder)
    full = os.path.join(folder, short)
    ultra_verbose("Full path of", short, "is", full)
    return full


def read_config():
    # Write/prepare first time configuration.
    config = config_filename(_CONFIG_FILE)
    if not os.path.exists(config):
        very_verbose("First time initializization of config file:", config)
        py_path = os.path.realpath(__file__)
        config_path = py_path[:-3] + ".default"
        ultra_verbose("Reading from template:", py_path)
        with open(config_path, "r") as template:
            default_config = template.read()

        with open(config, "w") as f:
            f.write(default_config)

    # Read config file
    with open(config, "r") as f:
        section = "(none)"
        for l in f.readlines():
            l = l.split(";")[0].split("#")[0].strip()
            if len(l) == 0:
                continue
            if l[0] == "[" and l[-1] == "]":
                section = l[1:-1]
                continue
            key, value = clean_split(l)
            if section == _SECTION_GROUPS or section == _SECTION_CUSTOM:
                for env in value.split(","):
                    ultra_verbose(_SECTION_GROUPS, key, env)
                    _groups[key].append(env.strip())
            elif section == _SECTION_HIGHLIGHT:
                for env in value.split(","):
                    ultra_verbose(_SECTION_HIGHLIGHT, env, key)
                    _highlight[env.strip()] = key
            elif section.startswith(_SECTION_PROFILE_START):
                profile = section[len(_SECTION_PROFILE_START):]
                ultra_verbose(_SECTION_PROFILE_START, profile, key, value)
                _profiles[profile][key] = value
            else:
                very_verbose("Ignoring config item:", section, key, value)

    if _verbose_level > 1:
        for p in sorted(_profiles.keys()):
            ultra_verbose("profile", p)
            for k, v in _profiles[p].items():
                ultra_verbose("  {k}={v}", k=k, v=v)

    # Read default environment file
    default_file = config_filename(_DEFAULT_FILE)
    if os.path.exists(default_file):
        with open(default_file, "r") as f:
            for l in f.readlines():
                l = l.strip()   # Removing trailing LF
                key, value = l.split("=", 1)
                ultra_verbose("reading default env", key, "=", value, ".")
                _default[key] = value


def save_default():
    default = config_filename(_DEFAULT_FILE)
    with open(default, "w") as f:
        for k in sorted(os.environ.keys()):
            f.write("{}={}\n".format(k, os.environ.get(k)))


def clear_default():
    default = config_filename(_DEFAULT_FILE)
    os.remove(default)


def main(arguments):
    read_config()

    if arguments.clear_default:
        clear_default()
        verbose("Default cleared")
        return 0
    elif arguments.default:
        save_default()
        verbose("Current environment set as default")
        return 0

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
                output_key(k, maxlen, password=arguments.show_password)
        else:
            not_displayed_group.append(group)

    not_currently_set = set(_default.keys()) - set(os.environ.keys())
    if not_currently_set:
        output_group("Removed from current env (unset)")
        for k in sorted(not_currently_set):
            output_key(k, maxlen)

    if len(not_displayed_group):
        output_group("Hidden groups: {}".format(" ".join(not_displayed_group)))

    active_profiles = []
    inactive_profiles = []
    for p in sorted(_profiles.keys()):
        ultra_verbose("profile", p)
        active = True
        for k, v in _profiles[p].items():
            if k not in os.environ or os.environ[k] != v:
                active = False
                break
        if active:
            active_profiles.append(p)
        else:
            inactive_profiles.append(p)

    output_profiles(active_profiles, inactive_profiles)

    # TODO: call shell_eval("export ...")

    return 0


if __name__ == "__main__":
    redirect_stdout()
    parser = argparse.ArgumentParser(description="Manage your environment with Envirou!")

    parser.add_argument("-v", "--verbose", action="count", default=0,
                        help="Increase output verbosity")

    parser.add_argument("-a", "--all", dest="all", action="store_true", help="Show all groups")
    parser.add_argument("-p", "--show-password", action="store_true", help="Clear out default")
    parser.add_argument("--no-all", dest="all", action="store_false", help="Don't show hidden groups (default)")
    parser.add_argument("--default", action="store_true", help="Set current env as default")
    parser.add_argument("--clear-default", action="store_true", help="Clear out default")

    parser.add_argument("group", nargs="*", help="Display named group(s)")

    args = parser.parse_args()
    _verbose_level = args.verbose
    exit(main(args))
