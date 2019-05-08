<h1>🖥&nbsp;&nbsp;am.ca-server&nbsp;&nbsp;📬</h1>
<i>My personal server that I use for www.alexmontague.ca</i>

![](https://i.imgur.com/Xn3AuBS.jpg)

- Currently this server is hosted on a raspberry pi, running 24/7 as my personal API and email server!
- I wanted a server that I could completely customize to my current needs, and have the scalability to fulfil my future needs!
- The API can be accessed at `api.alexmontague.ca` try pinging `/resume` for a JSON representation of my resume!
- Using this server as an opportunity to learn GoLang and networking stuff!

### How to Run
_If you want to spin up a similar Golang API, heres how you run mine_
1) Clone the repo
2) Make sure Go is installed
3) I suggest also having a session manager installed like [tmux](https://github.com/tmux/tmux) to persist sessions after sshing into the server (as I leave the pi running 24/7)
4) SSH into your server if applicable
5) Start a new session `tmux new -s API`
6) Run the server using `go run main.go`
7) The server will log endpoint info to stdout

*_The following steps are configuring the server as it runs on my personal network_*

### Create a Reverse Proxy
_I use nginx as a reverse proxy so I can direct requests from my default ip to the server running on my specific port_

1) Install [nginx](https://www.nginx.com/)
2) Start nginx and restart on reboots `sudo systemctl start nginx && sudo systemctl enable nginx`
3) Create an nginx config file in `/etc/nginx/conf.d` named whatever you want. I have this config named `goApi.conf`
4) In the config file paste this:
```
server {
  listen 80;
  listen [::]:80;

  server_name <EXTERNAL_DOMAIN>;

  location / {
      proxy_pass http://localhost:<PORT>/;
  }
}
```
Replace `<EXTERNAL_DOMAIN>` with your server's domain or ip <br/>
Replace `<PORT>` with the local port your server is running on. In my case it is `8080`<br />

5) Test the config file to make sure there are no errors: `sudo nginx -t`
6) If everything looks good, restart nginx `sudo nginx -s reload`


### Running the Cronjob
_I also use a shell script to update my api domain record, as it is hosted through my dynamic home external ip. Here is how you set that up!_
* *Note: This implementation is for GoDaddy Domains. If your domain provider has an API, you would need to make a few changes*  
1) Create a GoDaddy API Key
2) Store the credentials in `/etc/environment` as `GODADDY_KEY` and `GODADDY_SECRET` respectfully so the cron can access these environment vars
3) Set up the cron using `crontab -e` and adding the cron `*/10 * * * * ~/go/src/am.ca-server/updateIP.sh >> ~/Documents/cronAPI.log 2>&1`<br />
3.b) You do not need to redirect output if you do not want to log anything
4) If you did set up logging and wish to stream it in another tmux window, use `tail -f ~/Documents/cronAPI.log`
