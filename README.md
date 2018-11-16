# go-twitter-to-es

Use your "Likes" like a crappy bookmark service?
Are you trying to find stuff in your likes, and actually never do?
This small program will fill that gap in your life.


# Building and running
You need go for this!
You also need ElasticSearch and maybe Kibana if you like searching in a GUI.

If your ElasticSearch is not running on localhost you have to change this particular setting in ```src/main.go```

```
go get github.com/anaskhan96/soup
go build -o twitter-to-es ./src
./twitter-to-es -location=./liked_tweets.csv
```

