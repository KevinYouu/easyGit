#!/bin/bash
# Created by KevinYouu on 2024-03-24 14:30:11
# Description: install easyGit
# Github: https://github.com/KevinYouu/easyGit

repo="KevinYouu/easyGit"

# Check if git is installed
if ! command -v git &>/dev/null; then
    echo "git is not installed. Please install git first."
    echo "https://git-scm.com/downloads"
    exit 1
fi

version=$(curl -s https://api.github.com/repos/"$repo"/releases/latest | grep 'tag_name' | cut -d\" -f4)
echo "latest version: $version"

package_name=""
install_dir="/usr/local/bin"

function systemCheck() {
    system=$(uname -s)
    arch=$(uname -m)

    case "$system" in
    Darwin)
        if [ "$arch" == "x86_64" ]; then
            package_name="darwin_amd64"
            install_dir="/usr/local/bin"
        elif [ "$arch" == "arm64" ]; then
            package_name="darwin_arm64"
            install_dir="/opt/homebrew/bin"
        else
            echo "Your system is not supported."
            exit 1
        fi
        ;;
    Linux)
        if [ "$arch" == "x86_64" ]; then
            package_name="linux_amd64"
        elif [ "$arch" == "aarch64" ]; then
            package_name="linux_arm64"
        else
            echo "Your system is not supported."
            exit 1
        fi
        ;;
    *)
        echo "Your system is not supported."
        exit 1
        ;;
    esac
}

systemCheck
url=https://github.com/$repo/releases/download/$version/easyGit_"$version"_"$package_name".zip
file=easyGit_"$version"_"$package_name".zip

echo "Downloading $url"

if wget "$url"; then
    echo "Download successful"
else
    echo "Download failed"
    exit 1
fi

echo "Extracting and installing $file to $install_dir"

if sudo unzip -o "$file" -d "$install_dir"; then
    echo "Extraction successful"
else
    echo "Extraction failed"
    exit 1
fi

# Set permissions
if sudo chmod +x "$install_dir/easyGit"; then
    echo "Permissions set successfully"
else
    echo "Failed to set permissions"
    exit 1
fi

echo "Installation completed successfully"

rm "$file"

echo "Please restart your terminal or source your config manually."

# Initialize config
if [ ! -d "$install_dir/easyGit" ]; then
    "$install_dir/easyGit" init
    echo "Config initialized successfully"
fi
