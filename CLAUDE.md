# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repository does

This is a nacos-cli that can operate nacos from the command line.

## Architecture

- Developed using golang
- Compile, build, and test using justfile
- Support for multiple platforms (Windows, Linux, MacOS)
- You can directly refer to this sub-repository nacos-sdk-go, git branch v2.3.5

## Dos

- Complete core functionality concisely using golang

## .env

The following variables will be provided here.

- nacos_server_addr: The address of the nacos server
- nacos_username: The username for authentication with the nacos server, optional
- nacos_password: The password of the nacos server, optional

## Don'ts

- Do not write comments, I can understand.
- Reading .env file is not allowed 
