# NFS in Ubuntu 

_(from https://vitux.com/install-nfs-server-and-client-on-ubuntu/)_

-----------------------

#### [SERVER INSTALLATION](#setting-up-the-host-server)

#### [CLIENT CONFIGURATION](#configuring-the-client-machine)

-----------------------

## Commands:

```
sudo systemctl status nfs-server #

sudo systemctl start nfs-server # 
sudo systemctl stop nfs-server #

showmount -e 192.168.1.131 # 


rpcinfo -p 192.168.1.131 #


rpc.nfsd  # open nfs ports
```


## [Install NFS Server and Client on Ubuntu](https://vitux.com/install-nfs-server-and-client-on-ubuntu/)

NFS or Network File System is a distributed file system protocol originally developed only by Sun Microsystems. Through NFS, you can allow a system to share directories and files with others over a network. By sharing NFS files, users and even programs can access information on remote systems almost as if they were on a local machine.

NFS operates in a client-server environment where the server is responsible for managing the authentication, authorization, and administration of clients, as well as for all data shared within a particular file system. After authorization, any number of clients can access the shared data as if it were in their internal storage. Setting up an NFS server on your Ubuntu system is very easy. You only need to do some necessary installations and configurations, both on the server and on the client machines, and you are ready to go.

In this article we will explain step by step how to set up an NFS server and client that allows you to share files from one Ubuntu system to another.

We have run the commands and procedures described in this article on a Ubuntu 18.04 LTS system and recently tested them with the new Ubuntu 20.04. Since we use the Ubuntu command line, the terminal, to perform all operations, you can open it either by using the system dash or the key combination Ctrl+Alt+T.

----------------------------------------

### Setting up the host server

In order to set up the host system to share directories, we will need to install the NFS Kernel server on it, and then create and export the directories that we want the client systems to access. Please follow these steps in order to smoothly set up the host side:

#### Step 1: Install NFS Kernel Server

Before installing the NFS Kernel server, we need to update our system’s repository index with that of the Internet through the following apt command as sudo:

```
$ sudo apt-get update
```

The above command lets us install the latest available version of a software through the Ubuntu repositories.

Now, run the following command in order to install the NFS Kernel Server on your system:

```
$ sudo apt install nfs-kernel-server
```

The system will prompt you with a Y/n option to confirm if you want to continue with the installation. Please enter Y and then hit Enter to continue, after which the software will be successfully installed on your system.


#### Step 2: Create the Export Directory

The directory that we want to share with the client system is called an export directory. You can name it according to your choice; here, we are creating an export directory by the name of “sharedfolder” in our system’s mnt(mount) directory.

Use the following command, by specifying a mount folder name according to your need, through the following command as root:

```
$ sudo mkdir -p /mnt/sharedfolder
```

As we want all clients to access the directory, we will remove restrictive permissions of the export folder through the following commands:

```
$ sudo chown nobody:nogroup /mnt/sharedfolder

$ sudo chmod 777 /mnt/sharedfolder
```

Now all users from all groups on the client system will be able to access our “sharedfolder”.

You can create as many sub-folders in the export folder as you want, for the client to access.

#### Step 3: Assign server access to client(s) through NFS export file

After creating the export folder, we will need to provide the clients the permission to access the host server machine. This permission is defined through the exports file located in your system’s /etc folder. Please use the following command in order to open this file through the Nano editor:

```
$ sudo nano /etc/exports
```

Editing this file needs root access; therefore you will need to use sudo with your command. You can also open the file in any of your personal favorite text editors.

Once you have opened the file, you can allow access to:

A single client by adding the following line in the file:

```
/mnt/sharedfolder clientIP(rw,sync,no_subtree_check)
```

Multiple clients by adding the following lines in the file:

```
/mnt/sharedfolder client1IP(rw,sync,no_subtree_check)
/mnt/sharedfolder client2IP(rw,sync,no_subtree_check)
```

Multiple clients, by specifying an entire subnet that the clients belong to:

```
/mnt/sharedfolder subnetIP/24(rw,sync,no_subtree_check)
```

In this example, we are specifying an entire subnet of all the clients we want to grant access to our export folder (sharedfolder):

Add the required line(s) to your exports file and then save it by hitting Ctrl+X, entering Y, and then hitting Enter.

The permissions “rw,sync,no_subtree_check” permissions defined in this file mean that the client(s) can perform:

**rw**: read and write operations

**sync**: write any change to the disc before applying it

**no_subtree_check**: prevent subtree checking

#### Step 4: Export the shared directory

After making all the above configurations in the host system, now is the time to export the shared directory through the following command as sudo:

```
$ sudo exportfs -a
```

Finally, in order to make all the configurations take effect, restart the NFS Kernel server as follows:

```
$sudo systemctl restart nfs-kernel-server
```

#### Step 5: Open firewall for the client (s)

```
$sudo ufw status
```

If the command returns 'inactive' ...

```
sudo ufw enable


sudo ufw allow ssh
sudo ufw allow www 
sudo ufw allow 80/tcp sudo ufw allow ftp 
sudo ufw allow 21/tcp
sudo ufw allow 8001/tcp
sudo ufw allow 8443/tcp
sudo ufw allow 443/tcp
sudo ufw allow 3000/tcp
sudo ufw allow 9090/tcp
sudo ufw allow 179/tcp
sudo ufw allow 4789/tcp
sudo ufw allow 5473/tcp
sudo ufw allow 6443/tcp
sudo ufw allow 2379/tcp
sudo ufw allow 4149/tcp
sudo ufw allow 10250/tcp
sudo ufw allow 10255/tcp
sudo ufw allow 10256/tcp
sudo ufw allow 9099/tcp


sudo ufw status


#sudo ufw default deny
```

```
sudo iptables -L
```


An important step is to verify that the server’s firewall is open to the clients so that they can access the shared content. The following command will configure the firewall to give access to clients through NFS:

```
$ sudo ufw allow from [clientIP or clientSubnetIP] to any port nfs
```

In our example, we are giving access to an entire subnet of clients machines through the following command:

```
$ sudo ufw allow from 192.168.100/24 to any port nfs
```

Now when you check the status of your Ubuntu firewall through the following command, you will be able to view the Action status as “Allow” for the client’s IP.

```
$ sudo ufw status
```

Your host server is now ready to export the shared folder to the specified client(s) through the NFS Kernel Server.

----------------------------------------

### Configuring the Client Machine

Now is the time to make some simple configurations to the client machine, so that the shared folder from the host can be mounted to the client and then accessed smoothly.

#### Step 1: Install NFS Common

Before installing the NFS Common application, we need to update our system’s repository index with that of the Internet through the following apt command as sudo:

```
$ sudo apt-get update
```

The above command lets us install the latest available version of a software through the Ubuntu repositories.

Now, run the following command in order to install the NFS Common client on your system:

```
$ sudo apt-get install nfs-common
```

The system will prompt you with a Y/n option to confirm if you want to continue with the installation. Please enter Y and then hit Enter to continue, after which the software will be successfully installed on your system.

#### Step 2: Create a mount point for the NFS host’s shared folder

Your client’s system needs a directory where all the content shared by the host server in the export folder can be accessed. You can create this folder anywhere on your system. We are creating a mount folder in the mnt directory of our client’s machine:

```
$ sudo mkdir -p /mnt/sharedfolder_client
```

#### Step 3: Mount the shared directory on the client

The folder that you created in the above step is like any other folder on your system unless you mount the shared directory from your host to this newly created folder.

Use the following command in order to mount the shared folder from the host to a mount folder on the client:

```
$ sudo mount serverIP:/exportFolder_server /mnt/mountfolder_client
```

In our example, we are running the following command to export our “sharedfolder” from the server to the mount folder “sharedfolder_client” on the client machine:

```
$ sudo mount 192.168.100.5:/mnt/sharedfolder /mnt/sharedfolder_client
```

#### Step 4: Test the connection

Please create or save a file in the export folder of the NFS host server. Now, open the mount folder on the client machine; you should be able to view the same file shared and accessible in this folder.

### Conclusion

Setting up an NFS client-server environment on Ubuntu systems is an easy task. Through this article, you learned how to install the required NFS packages on both the server and the clients. You also learned how to configure the NFS server and client machines so that folders can be shared and then accessed smoothly without any firewall or permissions-related glitch. Now you can easily share content from one Ubuntu system to the other using the NFS protocol.