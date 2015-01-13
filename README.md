Stateless HTTP load balancer, that distributes requests round robin to the connected backend servers.

#Install
####Install binary
```
go install github.com/nwolber/proxy
```

####Install library: [Godoc](https://godoc.org/github.com/nwolber/proxy/rrproxy)
```
go install github.com/nwolber/proxy/rrproxy
```

#Options
Option | Example | Explanation
--- | --- | ---
-ep [*address*]:*port* | ```-ep :8080``` | Endpoint to listen on (Default: :80 or :443, depending on -key and -cert parameters)
-path /*path* | ```/myService?param=value``` | Frontend path to listen on, may include querry parameters (Default: /)
-cert *file* | ```-cert cert.pem``` | Path to a PEM-encoded certificate file (Optional)
-key *file* | ```-key key.pem``` | Path to a PEM-encoded key file (Optional)

#Usage
####HTTP
Listen for HTTP connections and redirect them to alice and bob.
```
proxy -path /funkyService http://alice/boringService http://bob/boringService
```
####HTTPS
Listen for HTTPS connections on port 8080 and redirect them to alice and bob.
```
proxy -cert="cert.pem" -key="key.pem" -ep :8080 http://alice http://bob
