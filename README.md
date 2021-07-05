# RouterTCPLimit

### 文件描述符限制
```
# cat <<EOF >> /etc/security/limits.conf
* hard nofile 97816
* soft nofile 97816

EOF

# ulimit -n
```

### 可用端口号
```
# cat <<EOF >> /etc/sysctl.conf
net.ipv4.ip_local_port_range = 10000 65530

EOF
```
### 建立的连接数查看
```
# netstat -n | awk '/^tcp/ {++state[$NF]} END {for(key in state) print key,"\t",state[key]}'
```
