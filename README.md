# azaffinity

## Name

*azaffinity* - adds the availability zone (AZ) name as a subdomain prefix to CNAME records resolving the DNS name for internal AWS ALB.

## Description

This plugin can be used only once.

## Syntax

```
azaffinity {
    SUBNET AZ
}
```

- **SUBNET** subnet CIDR range to check the client source IP
- **AZ** the AZ subdomain prefix to be added to the CNAME record

## Examples

To demonstrate the usage of plugin azaffinity, here we provide some typical examples.

~~~ corefile
. {
    azaffinity {
        10.0.0.0/24 ap-southeast-1a
    }
}
~~~

input CNAME: internal-k8s-123456.ap-southeast-1.elb.amazonaws.com
client source IP: 10.0.0.123
output CNAME: ap-southeast-1a.internal-k8s-123456.ap-southeast-1.elb.amazonaws.com