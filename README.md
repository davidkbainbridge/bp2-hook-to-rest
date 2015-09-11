# ![](bp_logo.png#top) BluePlanet bp2-hook-to-rest [![blueplanetutil](https://img.shields.io/badge/blueplanet-utility-blue.svg)]()
This utility can be used by BluePlanet microservices to redirect platform hook invocations to a REST call into the microservice.

This utility scrubs the environment variables for variables that begin with `BP_` and converts the resultant set into a JSON object. Further, any variable that ends in `_DATA` or `_CONFIG` is assumed to be JSON data marshaled into a string; thus this utility converts these values back to JSON structures.

The resultant JSON structure is POSTed to the URL `http://127.0.0.1/api/v1/hook/<name>` where `<name>` is replaced with the name of the executable being invoked, i.e. `southbound-update` would results in a post to `http://127.0.0.1/api/v1/hook/southbound-update`.

> Note: The use of this hook implies that the hook and application REST interfaces are responsive to calls from the platform. This may not be the case, i.e. BP starts a service and a consumer of that service, it then attempts to update the consumer of the service with connectivity information to the service, if that call executes the hook which attempts to connect to the consumer via REST and that REST call fails because the consumer is not yet listening on the REST interface then things are broken.

> It appears that a common solution to this is tha the hook script writes to the local file system and the consumer essentially reads those updates. This is a poor man's persistent message queue, but it addresses the issue of dependency injection after the fact, which is the issue at the core.

### Usage
To use this utility copy the executable into you `/bp2/hooks` directory within your docker container. Then symbolically link this file to the various BluePlanet hooks for which you would like to use it. For example, to use this for the southbound-update hook you would issue a command similar to

    ln -s /bp2/hooks/hook-to-rest /bp2/hooks/southbound-update

Afterward, a listing on the directory might look something like the following:

    # ls -l
    total 6844
    -rwxr-xr-x    1 root     root       7008112 Sep 10 22:16 hook-to-rest
    lrwxrwxrwx    1 root     root            23 Sep 10 22:17 southbound-update -> /bp2/hooks/hook-to-rest

### Customization
As this utility is meant to work with the BluePlanet platform, configuration of this utility's behavior is accomplished via environment variables set on the container.

> `BP_HOOK_URL_REDIRECT_<name>` - used to override the default target URL to which the data is posted, where `<name>` is replaces with the name of the executable in all uppercase characters and `-` (_dashes_) replaced with `_` (_underbars_), i.e. `southbound-update` would be converted to `BP_HOOK_URL_REDIRECT_SOUTHBOUND_UPDATE`.

> `BP_HOOK_DATA_SUFFIX_LIST` - used to override the list of suffixes that are checked against the environment variables when looking for embedded JSON data. By default this list containts `_DATA` and `_CONFIG`. That value of this variable should be a comma separated list; thus for the default suffixes this value would be `"_DATA, _CONFIG"`.

### Command Line Options
This utility supports the following command line options that can help with debugging or understanding the behavior of the utility.

> `-n` - do not execute the `HTTP POST`, but instead only display the target URL of the `POST` and the data that would be passed along with the `POST`.

> `-v` - produce verbose logging

### Example - Southbound Update
An example of a southbound update invocation might include the setting of the following environment variables (_note_: reformatted for readability):

    BP_HOOK_SOUTHBOUND_DATA='[
        {
            "name": "rabbitmq",
            "url": "amqpc://guest:guest@172.17.0.156:5672",
            "ip": "172.17.0.156",
            "type": "amqp",
            "port": "5672"
        }
    ]'
    BP_HOOK_NAME=southbound-update
    BP_HOOK_SOUTHBOUND_INTERFACE=rabbitmq

The utility would, by default `POST` the following data to `http://127.0.0.1/api/v1/hook/southbound-update` (_note_: reformatted for readability):

    {
        "BP_HOOK_NAME":"southbound-update",
        "BP_HOOK_SOUTHBOUND_DATA":[
            {
                "ip":"172.17.0.156",
                "name":"rabbitmq",
                "port":"5672",
                "type":"amqp",
                "url":"amqpc://guest:guest@172.17.0.156:5672"
            }
        ],
        "BP_HOOK_SOUTHBOUND_INTERFACE":"rabbitmq"
    }
