## "Objekte im Kreuzverhör" e-learning application

### Frontend

frontend is awaited in folder static/ 

### Forwarding ports using iptables

To redirect http/https:
```bash
iptables -t nat -A PREROUTING -i eth0 -p tcp --dport 80 -j REDIRECT --to-port 8080
iptables -t nat -A PREROUTING -i eth0 -p tcp --dport 443 -j REDIRECT --to-port 1443
```

To redirect https (loopback, localhost for development):
```bash
iptables -t nat -A OUTPUT -o lo -p tcp --dport 443 -j REDIRECT --to-port 1443
```
