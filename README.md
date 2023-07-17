# automaed_ssh
I created this tool for the purpose of automating the setup of Linux Servers and various other tasks.
The idea was to keep it really simple and use SSH. I do highly encourage people to please read the code.

To compile, I did use go 1.20 which is what I installed on my Debian server.

You can compile this program by running:

```
git clone https://github.com/f0rg-02/automaed_ssh
cd automaed_ssh && go build
```

To run, the program only takes one argument which is the yaml file.

```
./auto_ssh -f config.yaml
```

If there are any issues that are thrown, please file an issue, but try to troubleshoot. I am aware that
YAML errors are vague and a pita to understandd sometimes especially for newer people.

The config file is in YAML and is very simple and straightforward. This is what is used by the tool to connect
and execute commands and various other tasks.

Example:

```
server: "ssh_server"
port: "22"
username: "user"
sleep: 0
key_path: "key_path_of_the_client"

commands: [ "whoami", "ls -lah",  "ip a" ]

apt_update: true
apt_upgrade: true

apt_packages: [ "vagrant", "pipx", "p0f" ]

upload_files: true
files:
  file:
    - source: "test.txt"
      destination: "/home/user/test_copy.txt"

    - source: "test2.txt"
      destination: "/home/user/test_copy2.txt"
```

The `apt_update`, `apt_upgrade`, and `upload_files` must be set to true if you want to run those specific tasks.
Hopefully soon I can also code updating and upgrading on Red Hat, CentOS, and possibly other Linux distrobutions, but
for now, I have only tested on Debian Book Worm.

For any sudo commands, I specified it so it asks for user password via the term package on the client side and the command
that is run is:

```_, err := client.Run("echo " + password + "| sudo -S apt update")```

This was due to the limitations of the ssh library that I used and it is not recommended to connect to SSH as root or set
NOPASSWD in sudoers. It is a little janky, but it works and works well.

For the SSH keys, generate a pair on both systems and specify the private key of the client aka the computer that this tool is running on.

You can do both with `ssh-keygen` and a command is: `ssh-keygen -b 2048 -t rsa`

RSA-2048 is what I recommend minimum.

On the client run `ssh-copy-id user@server` to copy the ssh public key to your client.

#### TODO: Write a simple util key to generate ssh keys and copy them to the client all on the client end (Will get to this once I find time).
