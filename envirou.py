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
_CONFIG_PATH = "~/.config/envirou"
_CONFIG_FILE = "config.ini"
_CONFIG_DEFAULT_FILE = "config.default.ini"
_DEFAULT_FILE = "default"
_SECTION_SETTINGS = "settings"
_SECTION_GROUPS = "groups"
_SECTION_PROFILES = "profiles"
_SECTION_CUSTOM = "custom"
_SECTION_HIGHLIGHT = "highlight"
_SECTION_PROFILE_START = "profile:"
_HIGHLIGHT_PASSWORD = "password"
_CONFIG_DIFFERENCES = "differences"
_SETTINGS_QUIET = "quiet"
_NA_GROUP = "na"

_verbose_level = 0
_environ = {}
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
    out = fmt.format(**kwargs)
    very_verbose(" [eval] " + out, *args)
    print(out, *args, file=_stdout, end="")
    _stdout.flush()


def ultra_verbose(fmt, *args, **kwargs):
    if _verbose_level > 1:
        output(fmt, *args, **kwargs)


def very_verbose(fmt, *args, **kwargs):
    if _verbose_level > 0:
        output(fmt, *args, **kwargs)


def output(fmt, *args, **kwargs):
    kwargs["file"] = sys.stderr
    kwargs.update(**_CONSOLE_COLORS)
    output = fmt.format(**kwargs)
    print(output, *args)


def color_wrap(s, color):
    return "{c-" + color + "}" + s + "{c-end}"


def output_group(group):
    if group == _NA_GROUP:
        group += " (No Applicable group)"
    out = color_wrap("# {group}", color=_highlight.get(_SECTION_GROUPS, "magenta"))
    output(out, group=group)


def output_key(k, maxlen, no_diff=False, password=False):
    has_password = False
    fmt = "{key:<{maxlen}} {value}"
    value = os.environ.get(k, "")
    prefix = ""
    if _default:
        prefix = "  "
        if _default and ((k in _default and value != _default[k]) or k not in _default):
            if not no_diff:
                diff_color = _highlight.get(_CONFIG_DIFFERENCES, "red")
                prefix = color_wrap("* ", color=diff_color)
            else:
                prefix = "* "
    if k in _highlight:
        color = _highlight.get(k)
        if color == _HIGHLIGHT_PASSWORD:
            if not password:
                has_password = True
                value = "*" * len(value)
        else:
            fmt = color_wrap(fmt, color)
    output(prefix + fmt, key=k, value=value, maxlen=maxlen)
    return has_password


def output_profiles(active_profiles):
    active = []
    inactive = []
    for p in sorted(_profiles.keys()):
        if p in active_profiles:
            active.append(p)
        else:
            inactive.append(p)

    def_color = _highlight.get(_SECTION_GROUPS, "magenta")
    active_color = _highlight.get(_SECTION_PROFILES, "yellow")
    active_str = color_wrap(", ", def_color).join([color_wrap(p, active_color) for p in active])
    inactive_str = ", ".join(inactive)
    s = ""
    if active:
        s = color_wrap("# Profiles: ", def_color) + active_str

    if inactive and active:
        s += color_wrap(" (inactive: {})  [-p NAME to activate]".format(
            inactive_str), def_color)
    elif inactive:
        s = color_wrap("# Inactive profiles: {}  [-p NAME to activate]".format(
            inactive_str), def_color)

    if s:
        output(s)


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


def read_environ():
    global _environ
    if sys.stdin.isatty():
        _environ = os.environ
    else:
        for line in sys.stdin.readlines():
            k, v = clean_split(line)
            _environ[k] = v


def read_config():
    # Write/prepare first time configuration.
    config = config_filename(_CONFIG_FILE)
    if not os.path.exists(config):
        very_verbose("First time initialization of config file:", config)
        py_path = os.path.dirname(__file__)
        config_path = os.path.join(py_path, _CONFIG_DEFAULT_FILE)
        ultra_verbose("Reading from template:", py_path)
        with open(config_path, "r") as template:
            default_config = template.read()

        with open(config, "w") as f:
            f.write(default_config)
    else:
        very_verbose("Reading existing config file:", config)

    # Read config file
    with open(config, "r") as f:
        section = "(none)"
        for l in f.readlines():
            raw = l.split(";")[0].split("#")[0]
            l = raw.strip()
            if len(l) == 0:
                continue
            if l[0] == "[" and l[-1] == "]":
                section = l[1:-1]
                continue
            if "=" in l:
                key, value = clean_split(l)
            else:
                key = l
                value = None

            if section == _SECTION_SETTINGS:
                for env in value.split(","):
                    ultra_verbose(_SECTION_SETTINGS, key, env)
                    if key == _SETTINGS_QUIET:
                        global _verbose_level
                        _verbose_level -= int(value)
            elif section == _SECTION_GROUPS or section == _SECTION_CUSTOM:
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


def get_active_profiles():
    result = set()
    for p in _profiles.keys():
        ultra_verbose("profile", p)
        ultra_verbose(" ", _profiles[p])
        active = True
        for k, v in _profiles[p].items():
            ultra_verbose(" -> ", k, v,
                          _environ.get(k, "[not found]"))
            if v is None and k in _environ:
                ultra_verbose(
                    "not active (should not be there but is)")
                active = False
                break
            if v is not None and (k not in _environ or _environ[k] != v):
                ultra_verbose("not active (not equal)".format(
                    _environ.get(k, "[not found]"), v))
                active = False
                break
        if active:
            result.add(p)
    return result


def edit_config_file():
    if _environ.get("EDITOR", ""):
        shell_eval("$EDITOR", config_filename(_CONFIG_FILE))
        return 0
    else:
        output("Set your EDITOR env variable or edit file: ", config_filename(_CONFIG_FILE))
        return 1


def save_default():
    default = config_filename(_DEFAULT_FILE)
    with open(default, "w") as f:
        for k in sorted(_environ.keys()):
            f.write("{}={}\n".format(k, _environ.get(k)))
    output("Current environment set as default")
    return 0


def clear_default():
    default = config_filename(_DEFAULT_FILE)
    if os.path.exists(default):
        os.remove(default)
        output("Default cleared")
    else:
        output("No default environment set  [-s to set]")
    return 0


def changed_from_default():
    ignore_keys = set()
    for group in _groups:
        is_no_diff = (group[0:2] == "..")
        if is_no_diff:
            ignore_keys.update(_groups[group])

    ignored = []
    remove = []
    update = []
    for k, v in _environ.items():
        if k not in _default.keys():
            if k in ignore_keys:
                ignored.append(k)
            else:
                remove.append(k)
        elif v != _default[k]:
            if k in ignore_keys:
                ignored.append(k)
            else:
                update.append(k)

    add = []
    for k, v in _default.items():
        if k not in _environ.keys():
            if k in ignore_keys:
                ignored.append(k)
            else:
                add.append(k)

    if _verbose_level > 1:
        output("remove:", sorted(remove))
        output("update:", sorted(update))
        output("add:", sorted(add))
        output("ignored:", sorted(ignored))
    return remove, update, add, ignored


def output_no_change_required(ignored):
    if len(ignored) == 0:
        output("Nothing changed (run script / export VAR and run again)")
    else:
        output("Nothing important changed  [-dv for details]")
        very_verbose("Ignored changes in:", ", ".join(sorted(ignored)))


def reset_to_default():
    if not _default:
        output("No default environment set  [-s to set]")
        return 1

    remove, update, add, ignored = changed_from_default()

    if remove:
        very_verbose("Removing vars: " + ", ".join(remove))
    for k in remove:
        set_env_variable(k, None)

    if update:
        very_verbose("Updating vars: " + ", ".join(update))

    if add:
        very_verbose("Adding vars: " + ", ".join(add))

    for k in update + add:
        set_env_variable(k, _default[k])

    if remove or update or add:
        output("Environment reset to default")
    else:
        output_no_change_required(ignored)
    return 0


def diff_default():
    if not _default:
        output("No default environment set  [-s to set]")
        return 1

    # add <-> remove (since we are going the other way):
    add, update, remove, ignored = changed_from_default()

    if not (add or update or remove):
        output_no_change_required(ignored)
        return 0

    output_group("Changes required to get from default to current env")
    for k in sorted(update + add):
        output("export {k}={v}", k=k, v=shell_quote(os.environ.get(k, "")))
    for k in sorted(remove):
        output("unset {k}", k=k)

    output_group("If you want to create a new profile  [--edit]")
    output("[profile:new]")
    for k in sorted(update + add + remove):
        if k in remove:
            output("{k}", k=k)
        else:
            output("{k}={v}", k=k, v=os.environ.get(k, ""))

    return 0


def shell_quote(s):
    if s.find(" ") != -1 and s[0] != "\"" and s[0] != "'":
        return "\"{}\"".format(s)   # s.replace(" ", "\\ ")
    else:
        return s


def set_env_variable(k, v):
    ultra_verbose("set_env_variable", k, v)
    if v is None:
        shell_eval("unset {k};", k=k)
    else:
        shell_eval("export {k}={v};", k=k, v=shell_quote(v))


def activate_profile(p):
    if p in _profiles:
        for k, v in _profiles[p].items():
            set_env_variable(k, v)
        return True
    else:
        return False


def activate_all_profiles(profiles):
    very_verbose("Profiles to activate", repr(profiles))
    for p in profiles:
        if activate_profile(p):
            output("Profile '{p}' activated", p=p)
        else:
            output("Profile '{p}' not found", p=p)
            return 1
    return 0


def list_groups():
    for g in sorted(_groups.keys()):
        output_group(g)
    return 0


def main(arguments):
    read_config()

    read_environ()

    if arguments.edit:
        return edit_config_file()
    elif arguments.clear_default:
        return clear_default()
    elif arguments.set_default:
        return save_default()
    elif arguments.reset_to_default:
        return reset_to_default()
    elif arguments.diff_default:
        return diff_default()
    elif arguments.profile:
        return activate_all_profiles(arguments.profile)
    elif arguments.list:
        return list_groups()

    match_group = defaultdict(list)
    for name, keys in _groups.items():
        for k in sorted(keys):
            match_group[k].append(name)

    maxlen = max([len(k) for k in _environ.keys()])
    grouped = defaultdict(list)
    for k in sorted(_environ.keys()):
        matched_groups = match_group[k]

        if matched_groups:
            for group in matched_groups:
                grouped[group].append(k)
        else:
            grouped[_NA_GROUP].append(k)

    filter_groups = len(arguments.group) > 0
    not_displayed_group = []
    has_hidden_password = False
    for group in sorted(grouped.keys()):
        is_hidden = (group[0] == ".")
        is_no_diff = (group[0:2] == "..")
        if arguments.all or (filter_groups and group in arguments.group) or (not filter_groups and not is_hidden):
            output_group(group)
            for k in grouped[group]:
                if output_key(k, maxlen, no_diff=is_no_diff, password=arguments.show_password):
                    has_hidden_password = True
        else:
            not_displayed_group.append(group)

    not_currently_set = set(_default.keys()) - set(_environ.keys())
    if not_currently_set:
        output_group("Removed from current env (unset)")
        for k in sorted(not_currently_set):
            output_key(k, maxlen)

    if _verbose_level >= 0:
        # Suppressed if --quiet
        if has_hidden_password:
            output_group("Passwords hidden  [-w to show]")

        if len(not_displayed_group):
            output_group("Groups hidden: {}  [NAME or -a]".format(
                " ".join(not_displayed_group)))

    active_profiles = get_active_profiles()
    output_profiles(active_profiles)

    return 0


if __name__ == "__main__":
    redirect_stdout()
    parser = argparse.ArgumentParser(
        description="Manage your environment with Envirou! [ev]")

    parser.add_argument(
        "-w", "--show-password", action="store_true",
        help="Display passwords")
    parser.add_argument(
        "-e", "--edit", action="store_true",
        help="Edit Envirou configuration")
    parser.add_argument(
        "-v", "--verbose", action="count", default=0,
        help="Increase output verbosity")
    parser.add_argument(
        "-q", "--quiet", action="count", default=0,
        help="Suppress output verbosity")

    defaults = parser.add_argument_group(
        "Default env", "Compare environment with a fixed/default set")
    defaults.add_argument(
        "-s", "--set-default", action="store_true",
        help="Set current env as default")
    defaults.add_argument(
        "-c", "--clear-default", action="store_true",
        help="Clear out default")
    defaults.add_argument(
        "-d", "--diff-default", action="store_true",
        help="Show differences from default")
    defaults.add_argument(
        "-r", "--reset-to-default", action="store_true",
        help="Reset env to default")

    profiles = parser.add_argument_group(
        "Profiles", "Environment variable profiles")
    profiles.add_argument(
        "-p", "--profile", action="append",
        help="Activate profile")

    groups = parser.add_argument_group(
        "Groups", "Groups of environment variables")
    groups.add_argument(
        "group",
        nargs="*",
        help="Display group or groups")
    groups.add_argument(
        "-l", "--list", dest="list", action="store_true",
        help="List groups")
    groups.add_argument(
        "-a", "--all", dest="all", action="store_true",
        help="Show all groups")
    groups.add_argument(
        "-n", "--no-all", dest="all", action="store_false",
        help="Don't show hidden groups (default)")

    args = parser.parse_args()
    _verbose_level = args.verbose - args.quiet
    exit(main(args))
