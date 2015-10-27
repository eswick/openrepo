# Description
OpenRepo is an APT repository server, currently targeting jailbroken iOS devices. It can be used to (very easily) create a repository for use with Cydia or other APT frontends.

## Building
    go get github.com/blakesmith/ar
    go build
    
## Configuration
OpenRepo's configuration is an XML file stored at `/etc/openrepo/config.xml`. Currently, this file must be made manually.

    mkdir /etc/oprenrepo
    touch /etc/openrepo/config.xml

Here is an example config:

    <Config>
	    <HostPath>/</HostPath>
	    <PackagePath>/opt/openrepo/packages</PackagePath>
    	<Release>
		    <ReleaseEntry key="Origin">OpeniOS</ReleaseEntry>
	    	<ReleaseEntry key="Label">OpeniOS</ReleaseEntry>
    		<ReleaseEntry key="Suite">stable</ReleaseEntry>
    		<ReleaseEntry key="Version">0.9</ReleaseEntry>
    		<ReleaseEntry key="Codename">OpeniOS</ReleaseEntry>
    		<ReleaseEntry key="Architectures">iphoneos-arm</ReleaseEntry>
    		<ReleaseEntry key="Components">main</ReleaseEntry>
    		<ReleaseEntry key="Description">OpeniOS Repository</ReleaseEntry>
    	</Release>
    </Config>
    

* **Config**
  * Root XML key
* **HostPath**
  * The URL path at which the repository will be hosted (e.g. for "/cydia", the repository would be hosted at "apt.yoururl.com/cydia")
* **PackagePath**
  * The directory in the filesystem where your Debian packages are stored
* **Release**
  * A collection of **ReleaseEntry**
  * **ReleaseEntry**
    * The "key" attribute is the name of an entry's key in the Release file, the data is the entry's value

## Usage

To run OpenRepo:

    sudo openrepo

To host a package on your OpenRepo server, simply place your deb file in your configuration's **PackagePath** directory!
