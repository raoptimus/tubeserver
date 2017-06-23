openssl s_client -connect ...:443
curl -s http://...:6060/debug/vars | json queries | json -a