from __future__ import print_function
import argparse
import os
import sys
from collections import defaultdict

_CONSOLE_COLORS = {
    "[[c-end]]": "\033[0m",
    "[[c-bold]]": "\033[1m",
    "[[c-underline]]": "\033[4m",
    "[[c-red]]": "\033[31m",
    "[[c-green]]": "\033[32m",
    "[[c-yellow]]": "\033[33m",
    "[[c-blue]]": "\033[34m",
    "[[c-magenta]]": "\033[35m",
}

_CONFIG_ENV = "ENVIROU_HOME"
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
_HIGHLIGHT_PATH = "path"
_CONFIG_DIFFERENCES = "differences"
_SETTINGS_QUIET = "quiet"
_SETTINGS_SORT_KEYS = "sort_keys"
_SETTINGS_PATH_TILDE = "path_tilde"
_NA_GROUP = "(no group)"
_BASH_COMPLETION_SCRIPT = """_envirou_completions() {
    COMPREPLY=($(compgen -W "$(envirou --inactive-profiles 2>&1)" -- "${COMP_WORDS[${COMP_CWORD}]}"));
};
complete -F _envirou_completions ev;
complete -F _envirou_completions envirou;"""

_ZSH_COMPLETION_SCRIPT = (
    """compdef '_values $(envirou --inactive-profiles 2>&1)' ev; compdef envirou=ev;"""
)
_CONFIG_ESCAPES_REQUIRED = "\\\r\n\t"
_POSIX_ESCAPES_REQUIRED = "\\$`\"\n\r\t"
_ESCAPE_PAIRS = [
    ("\\", "\\\\"), ("$", "\\$"), ("`", "\\`"), ("\"", "\\\""),
    ("\n", "\\n"), ("\r", "\\r"), ("\t", "\\t")
]
_verbose_level = 1
_sort_keys = True
_use_tilde = True
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


def shell_eval(fmt, *arguments, **kwargs):
    out = fmt.format(**kwargs)
    if _verbose_level > 1:
        with open("envirou_shell_debug.txt", "a") as f:
            print(out, *arguments, file=f, end="")
    very_verbose("{noformat}", noformat=" ".join([" [eval] ", out]))
    print(out, *arguments, file=_stdout)
    _stdout.flush()


def ultra_verbose(fmt, *arguments, **kwargs):
    if _verbose_level > 1:
        output(fmt, *arguments, **kwargs)


def very_verbose(fmt, *arguments, **kwargs):
    if _verbose_level > 0:
        output(fmt, *arguments, **kwargs)


def display_additional(s):
    if _verbose_level >= 0:
        return s
    else:
        return ""


def expand_console_colors(s):
    result = s
    for k, v in _CONSOLE_COLORS.items():
        result = result.replace(k, v)
    return result


def output(fmt, *arguments, **kwargs):
    out = expand_console_colors(fmt).format(**kwargs)
    print(out, *arguments)


def color_wrap(s, color):
    return "[[c-" + color + "]]" + s + "[[c-end]]"


def output_group(group):
    out = color_wrap("# {group}", color=_highlight.get(_SECTION_GROUPS, "magenta"))
    output(out, group=group)


def output_key(key, maxlen, no_diff=False, password=False, unformatted=False):
    has_password = False
    fmt = "{key:<{maxlen}} {value}"
    value = os.environ.get(key, "")
    prefix = ""
    if unformatted:
        output(key + "=" + value)
    else:
        if _default:
            prefix = "  "
            if (key in _default and value != _default[key]) or key not in _default:
                if not no_diff:
                    diff_color = _highlight.get(_CONFIG_DIFFERENCES, "red")
                    prefix = color_wrap("* ", color=diff_color)
                else:
                    prefix = "* "
        if key in _highlight:
            color = _highlight.get(key)
            if color == _HIGHLIGHT_PASSWORD:
                if not password:
                    has_password = True
                    if len(value) >= 16:
                        # Display last four digits of keys.
                        mask = "*" * (len(value) - 4)
                        value = mask + value[-4:]
                    else:
                        value = "*" * len(value)
            elif color == _HIGHLIGHT_PATH:
                user_path = os.path.expanduser('~')
                new_value = list()
                for i, path in enumerate(value.split(os.pathsep)):
                    if _use_tilde:
                        path = path.replace(user_path, '~')
                    new_value.append(color_wrap(path, "end" if i % 2 == 0 else "underline"))
                value = os.pathsep.join(new_value)
            else:
                fmt = color_wrap(fmt, color)
        output(prefix + fmt, key=key, value=expand_console_colors(value), maxlen=maxlen)
    return has_password


def case_insensitive_sort(c):
    return sorted(c, key=lambda s: s.lower())


def output_profiles(active_profiles):
    def_color = _highlight.get(_SECTION_GROUPS, "magenta")
    active_color = _highlight.get(_SECTION_PROFILES, "yellow")
    s = ""
    for p in case_insensitive_sort(_profiles.keys()):
        if p in active_profiles:
            s += " " + color_wrap(p, active_color)
        else:
            s += " " + p

    if s:
        s = color_wrap("# Profiles:", def_color) + s
        s += color_wrap(display_additional(" [NAME to activate]"), def_color)

        output(s)


def clean_split(s, sep="="):
    k, v = s.split(sep, 1)
    return k.strip(), v.strip()


def escape(s, reverse, escapes_required):
    if s is None:
        return s
    result = []
    if reverse:
        skip_next = False
        for i, c in enumerate(s):
            if skip_next:
                skip_next = False
                continue
            if c == "\\" and i + 1 < len(s):
                match = s[i:i+2]
                for a, b in _ESCAPE_PAIRS:
                    if b == match:
                        result.append(a)
                        skip_next = True
                        break
            else:
                result.append(c)
    else:
        for c in s:
            if c in escapes_required:
                for a, b in _ESCAPE_PAIRS:
                    if c == a:
                        result.append(b)
                        break
            else:
                result.append(c)
    return "".join(result)
    

def escape_config(s, reverse=False):
    return escape(s, reverse, _CONFIG_ESCAPES_REQUIRED)


def escape_posix_shell(s, reverse=False):
    return escape(s, reverse, _POSIX_ESCAPES_REQUIRED)


def config_filename(short):
    """
    :param short: short name
    :return: Get full path to config filename.
    """
    if os.name == "nt":
        default_prefix = os.environ.get("APPDATA", "~")
    else:
        default_prefix = "~/.config"
    folder = os.environ.get(_CONFIG_ENV, os.path.join(default_prefix, "envirou"))
    folder = os.path.expanduser(folder)
    if not os.path.isdir(folder):
        very_verbose("Creating configuration folder:", folder)
        os.makedirs(folder)
    full = os.path.join(folder, short)
    ultra_verbose("Full path of", short, "is", full)
    return full


def read_environ():
    global _environ
    _environ = {}
    if sys.stdin.isatty():
        for k, v in os.environ.items():
            if os.name == "posix":
                _environ[k] = escape_config(v, reverse=True)
            else:
                _environ[k] = v
    else:
        for line in sys.stdin.readlines():
            output("Parsing env line:" + line)
            try:
                k, v = clean_split(line)
                _environ[k] = v
            except ValueError:
                ultra_verbose("Malformed env (linefeed in values?)")


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
        for line in f.readlines():
            line = line.strip()
            if len(line) == 0 or line[0] == ";" or line[0] == "#":
                continue
            if line[0] == "[" and line[-1] == "]":
                section = line[1:-1]
                continue
            if "=" in line:
                key, value = clean_split(line)
            else:
                key = line
                value = None

            if section == _SECTION_SETTINGS:
                global _verbose_level, _sort_keys, _use_tilde
                for env in value.split(","):
                    ultra_verbose(_SECTION_SETTINGS, key, env)
                    if key == _SETTINGS_QUIET:
                        _verbose_level -= int(value)
                    elif key == _SETTINGS_SORT_KEYS:
                        _sort_keys = int(value) > 0
                    elif key == _SETTINGS_PATH_TILDE:
                        _use_tilde = int(value) > 0
            elif section == _SECTION_GROUPS or section == _SECTION_CUSTOM:
                for env in value.split(","):
                    ultra_verbose(_SECTION_GROUPS, key, env)
                    _groups[key].append(env.strip())
            elif section == _SECTION_HIGHLIGHT:
                for env in value.split(","):
                    ultra_verbose(_SECTION_HIGHLIGHT, env, key)
                    _highlight[env.strip()] = key
            elif section.startswith(_SECTION_PROFILE_START):
                profile = section[len(_SECTION_PROFILE_START):].strip()
                ultra_verbose(_SECTION_PROFILE_START, profile, key, value, repr(escape_config(value, reverse=True)))
                _profiles[profile][key] = escape_config(value, reverse=True)
            else:
                very_verbose("Ignoring config item:", section, key, value)

    if _verbose_level > 1:
        for p in sorted(_profiles.keys()):
            ultra_verbose("profile", p)
            for k, v in _profiles[p].items():
                ultra_verbose("  {k}={v}", k=k, v=repr(v))

    # Read default environment file
    default_file = config_filename(_DEFAULT_FILE)
    if os.path.exists(default_file):
        with open(default_file, "r") as f:
            for line in f.readlines():
                line = line.strip()  # Removing trailing LF
                key, value = line.split("=", 1)
                ultra_verbose("reading default env", key, "=", value, ".")
                _default[key] = escape_config(value, reverse=True)


def add_to_config_file(lines):
    config = config_filename(_CONFIG_FILE)
    very_verbose("Adding to config file:\n" + "\n".join(lines))
    with open(config, "a") as f:
        f.writelines("\n".join(lines))


def get_profiles(inactive_only=False):
    result = set()
    for p in _profiles.keys():
        ultra_verbose("profile", p)
        ultra_verbose(" ", _profiles[p])
        active = True
        for k, v in _profiles[p].items():
            ultra_verbose(" -> ", repr(k), repr(v), repr(_environ.get(k, "[not found]")))
            if v is None and k in _environ:
                ultra_verbose("not active (should not be there but is)")
                active = False
                break
            if v is not None and (k not in _environ or _environ[k] != v):
                ultra_verbose("not active (not equal)", repr(_environ.get(k)), repr(v))
                active = False
                break
        ultra_verbose("profile", p, "is", "active" if active else "not active")
        if (active and (not inactive_only)) or ((not active) and inactive_only):
            result.add(p)
    return result


def edit_config_file():
    if _environ.get("EDITOR", ""):
        fn = config_filename(_CONFIG_FILE)
        output("Editing config file: ", fn)
        shell_eval(_environ["EDITOR"], fn)
        return 0
    else:
        output(
            "Set your EDITOR env variable or edit file: ", config_filename(_CONFIG_FILE)
        )
        return 1


def save_default():
    default = config_filename(_DEFAULT_FILE)
    with open(default, "w") as f:
        for k in sorted(_environ.keys()):
            f.write("{}={}\n".format(k, escape_config(_environ.get(k))))
    output("Current environment set as default")
    return 0


def clear_default():
    default = config_filename(_DEFAULT_FILE)
    if os.path.exists(default):
        os.remove(default)
        output("Default cleared")
    else:
        output(
            "No default environment set  {help}", help=display_additional("[-s to set]")
        )
    return 0


def glob_match(glob, match):
    if glob == match:
        return True
    if len(glob) > 0 and glob[-1] == "*":
        if match.startswith(glob[:-1]):
            return True
    return False


def changed_from_default():
    ignore_keys = set()
    for group in _groups:
        is_no_diff = group[0:2] == ".."
        if is_no_diff:
            ignore_keys.update(_groups[group])

    ignored = []
    remove = []
    update = []
    for k, v in _environ.items():
        if k not in _default.keys() or v != _default[k]:
            append = True
            for ignore in ignore_keys:
                if glob_match(ignore, k):
                    ignored.append(k)
                    append = False
                    break
            if append:
                if k not in _default.keys():
                    remove.append(k)
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
        output(
            "Nothing changed  {help}",
            help=display_additional("(run script / export VAR and run again)"),
        )
    else:
        output(
            "Nothing important changed  {help}",
            help=display_additional("[-dv for details]"),
        )
        very_verbose("Ignored changes in:", ", ".join(sorted(ignored)))


def reset_to_default():
    if not _default:
        output(
            "No default environment set  {help}", help=display_additional("[-s to set]")
        )
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
        output(
            "No default environment set  {help}", help=display_additional("[-s to set]")
        )
        return 1

    # add <-> remove (since we are going the other way):
    add, update, remove, ignored = changed_from_default()

    if not (add or update or remove):
        output_no_change_required(ignored)
        return 0

    output_group(
        "To get from default to current env  {help}".format(
            help=display_additional("[-n PROFILE_NAME for new profile]")
        )
    )
    for k in sorted(update + add + remove):
        output(set_env_variable_command(k, os.environ.get(k)))
 
    return 0


def new_profile(profile_name):
    if not _default:
        output(
            "No default environment set  {help}", help=display_additional("[-s to set]")
        )
        return 1

    if profile_name in _profiles:
        output_profiles(get_profiles())
        output(
            "Profile {profile_name} already exists. You need a new name.",
            profile_name=profile_name,
        )
        return 1

    # add <-> remove (since we are going the other way):
    add, update, remove, ignored = changed_from_default()

    if not (add or update or remove):
        output_no_change_required(ignored)
        return 0

    lines = list(["", ""])
    lines.append("[profile:{profile_name}]".format(profile_name=profile_name))
    for k in sorted(update + add + remove):
        if k in remove:
            lines.append("{k}".format(k=k))
        else:
            lines.append("{k}={v}".format(k=k, v=escape_config(os.environ.get(k, ""))))
    add_to_config_file(lines)

    output("Profile {profile} created", profile=profile_name)
    return 0


def set_env_variable(k, v):
    ultra_verbose("set_env_variable", k, v)
    shell_eval(set_env_variable_command(k, v))
    
    
def set_env_variable_command(k, v):
    if os.name == 'posix':
        if v is None:
            return "unset {k};".format(k=k)
        else:
            return "export {k}=\"{v}\";".format(k=k, v=escape_posix_shell(v))
    else:  # nt
        return "set {k}={v}".format(k=k, v=v or "")


def activate_profile(p):
    if p in _profiles:
        for k, v in _profiles[p].items():
            set_env_variable(k, v)
        return True
    else:
        return False


def activate_all_profiles(profiles):
    very_verbose("Profiles to activate", repr(profiles))
    active_color = _highlight.get(_SECTION_PROFILES, "yellow")
    for p in profiles:
        if activate_profile(p):
            output(
                "Envirou profile {p} activated".format(p=color_wrap(p, active_color))
            )
        else:
            output("Envirou profile {p} not found", p=p)
            return 1
    return 0


def list_active_profiles_colored():
    active_color = _highlight.get(_SECTION_PROFILES, "yellow")
    output(color_wrap(" ".join(case_insensitive_sort(get_profiles())), active_color))
    return 0


def list_profiles(inactive_only=False):
    output(" ".join(case_insensitive_sort(get_profiles(inactive_only))))
    return 0


def bash_completions():
    shell_eval("{noformat}", noformat=_BASH_COMPLETION_SCRIPT)
    return 0


def zsh_completions():
    shell_eval("{noformat}", noformat=_ZSH_COMPLETION_SCRIPT)
    return 0


def list_groups():
    for g in sorted(_groups.keys()):
        output_group(g)
    return 0


def should_display_group(arguments, group_name):
    filter_groups = len(arguments.group) > 0
    is_hidden = group_name[0] == "."
    if (
        arguments.all
        or (filter_groups and group_name in arguments.group)
        or (not filter_groups and not is_hidden)
    ):
        return True
    return False


def main(arguments):
    read_config()

    read_environ()

    if arguments.old_profile:
        # Temporary for backward compatibility
        arguments.profile.extend(arguments.old_profile)

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
    elif arguments.new_profile:
        return new_profile(arguments.new_profile)
    elif len(arguments.profile) > 0:
        return activate_all_profiles(arguments.profile)
    elif arguments.active_profiles_colored:
        return list_active_profiles_colored()
    elif arguments.active_profiles:
        return list_profiles()
    elif arguments.inactive_profiles:
        return list_profiles(inactive_only=True)
    elif arguments.list:
        return list_groups()
    elif arguments.zsh_completions:
        return zsh_completions()
    elif arguments.bash_completions:
        return bash_completions()

    remaining_environ = set(_environ.keys())
    match_group = defaultdict(list)
    for name, keys in _groups.items():
        for k in keys:
            for env_item in _environ.keys():
                if glob_match(k, env_item):
                    match_group[name].append(env_item)
                    remaining_environ.discard(env_item)
    if remaining_environ:
        match_group[_NA_GROUP] = sorted(remaining_environ)

    # Calculate maximum length of output string
    maxlen = 1
    for group in match_group.keys():
        if should_display_group(arguments, group):
            for k in match_group[group]:
                maxlen = max(maxlen, len(k))

    not_displayed_group = []
    has_hidden_password = False
    for group in sorted(match_group.keys()):
        if should_display_group(arguments, group):
            is_no_diff = group[0:2] == ".."
            output_group(group)
            keys = match_group[group]
            if _sort_keys:
                keys = sorted(keys)
            for k in keys:
                if output_key(
                    k, maxlen, no_diff=is_no_diff, password=arguments.show_password, unformatted=args.unformatted
                ):
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
            output_group(
                "Groups hidden: {}  [-g NAME or --all]".format(
                    " ".join(not_displayed_group)
                )
            )

    active_profiles = get_profiles()
    output_profiles(active_profiles)

    return 0


if __name__ == "__main__":
    redirect_stdout()
    parser = argparse.ArgumentParser(
        description="Manage your environment with Envirou! [ev]"
    )

    parser.add_argument(
        "-w", "--show-password", action="store_true", help="Display passwords"
    )
    parser.add_argument(
        "-u", "--unformatted", action="store_true", help="Display without fancy formatting"
    )
    parser.add_argument(
        "-e", "--edit", action="store_true", help="Edit Envirou configuration"
    )
    parser.add_argument(
        "-v", "--verbose", action="count", default=0, help="Increase output verbosity"
    )
    parser.add_argument(
        "-q", "--quiet", action="count", default=0, help="Suppress output verbosity"
    )

    profile_group = parser.add_argument_group(
        "Profiles", "Environment variable profiles"
    )
    profile_group.add_argument("profile", nargs="*", help="Activate profile(s)")
    profile_group.add_argument(
        "-t", "--active-profiles", action="store_true", help="List active profiles"
    )
    profile_group.add_argument(
        "-i", "--inactive-profiles", action="store_true", help="List inactive profiles"
    )
    profile_group.add_argument(
        "-p", "--profile", dest="old_profile", action="append", help=argparse.SUPPRESS
    )

    groups = parser.add_argument_group("Groups", "Groups of environment variables")
    groups.add_argument(
        "-a",
        "--all",
        dest="all",
        action="store_true",
        help="Show all (including hidden groups)",
    )
    groups.add_argument(
        "-g", "--group", default=[], action="append", help="Display group or groups"
    )
    groups.add_argument(
        "-l", "--list", dest="list", action="store_true", help="List groups"
    )

    defaults = parser.add_argument_group(
        "Default env", "Compare environment with a fixed/default set"
    )
    defaults.add_argument(
        "-s", "--set-default", action="store_true", help="Set current env as default"
    )
    defaults.add_argument(
        "-c", "--clear-default", action="store_true", help="Clear out default"
    )
    defaults.add_argument(
        "-d",
        "--diff-default",
        action="store_true",
        help="Show differences from default",
    )
    defaults.add_argument(
        "-n",
        "--new-profile",
        help="Create a new profile named NEW_PROFILE from differences from default",
    )
    defaults.add_argument(
        "-r", "--reset-to-default", action="store_true", help="Reset env to default"
    )

    scripting = parser.add_argument_group(
        "Scripting", "Helpful information for scripts"
    )
    scripting.add_argument(
        "--active-profiles-colored",
        action="store_true",
        help="List active profiles in color",
    )
    scripting.add_argument(
        "--zsh-completions", action="store_true", help="Enable zsh completions"
    )
    scripting.add_argument(
        "--bash-completions", action="store_true", help="Enable bash completions"
    )

    args = parser.parse_args()
    _verbose_level = args.verbose - args.quiet
    exit(main(args))
