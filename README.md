# RouterTCPLimit

### 文件描述符限制
```
# cat <<EOF >> /etc/security/limits.conf
* hard nofile 97816
* soft nofile 97816

EOF

# ulimit -n
```

