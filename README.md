![release](https://github.com/shuntaka9576/kanban/workflows/release/badge.svg)

# Kanban
A Simple terminal UI for GitHub Project :octocat:.

![gif](https://github.com/shuntaka9576/kanban/blob/master/doc/gif/kanban.gif?raw=true)

## Installation
```
brew tap shuntaka9576/tap
brew install shuntaka9576/tap/kanban
```

## Configuration
1. Refer to [Personal access tokens](https://github.com/settings/tokens). Please create an access token.
2. Please create the following configuration file.
```
$ cat ~/.config/kanban
github.com:
  - user: [GitHub userID]
    oauth_token: [GitHub access token]
```
3. Please set the following environment variables.
```
export LC_CTYPE=en_US.UTF-8
```

## Usage
In the case of this repository, you can use `kanban --repo shuntaka9576/kanban`.

```
$ kanban --help
GitHub Project Viewer

Usage:
  kanban [flags]

Flags:
      --help              Show help for command
  -R, --repo OWNER/REPO   Select another repository using the OWNER/REPO format
  -S, --search string     Search project name string[default first project]
```

## Limit
There are a few limitations. Please give us your feedback.

|category|name|limit|
|---|---|---|
|project|project|1(use search option)|
|project|columns|10|
|columns|cards|100|
|issue|labels|10|
|issue|assignees|10|

## Features
* [ ] Real time preview
* [ ] Support organization project bord
