<h1>🖥&nbsp;&nbsp;am.ca-server&nbsp;&nbsp;📬</h1>
<i>My personal server that I use for www.alexmontague.ca</i>

![](https://i.imgur.com/Xn3AuBS.jpg)

- Currently this server is hosted on a raspberry pi, running 24/7 as my personal API and email server!
- I wanted a server that I could completely customize to my current needs, and have the scalability to fulfil my future needs!
- Currently can be accessed at `api.alexmontague.ca:8088` try pinging `/resume` for a JSON representation of my resume!
- Using this server as an opportunity to learn GoLang and networking stuff while still being useful!

### How to Run
_If you want to spin up a similar Golang API, heres how you run mine_
1) Clone the repo
2) Make sure Go is installed
3) I suggest also having a session manager installed like [tmux](https://github.com/tmux/tmux) to persist sessions after sshing into the server (as I leave the pi running 24/7)
4) SSH into your server if applicable
5) Start a new session `tmux new -s API`
6) Run the server using `go run main.go`
7) The server will log endpoint info to stdout

### Running the Cronjob
_I also use a shell script to update my api domain record, as it is hosted on my dynamic home network. Here is how you set that up!_
* *Note: This implementation is for GoDaddy Domains. If your domain provider has an API, you would need to make a few changes*  
1) Create a GoDaddy API Key
2) Store the credentials in `/etc/environment` as `GODADDY_KEY` and `GODADDY_SECRET` respectfully so the cron can access these environment vars
3) Set up the cron using `crontab -e` and adding the cron `*/10 * * * * ~/go/src/am.ca-server/updateIP.sh >> ~/Documents/cronAPI.log 2>&1`<br />
3.b) You do not need to redirect output if you do not want to log anything
4) If you did set up logging and wish to stream it in another tmux window, use `tail -f ~/Documents/cronAPI.log`
