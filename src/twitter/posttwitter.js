const { TwitterApi } = require('twitter-api-v2');
const yaml = require("js-yaml");
const fs = require("fs");

const setting = yaml.safeLoad(fs.readFileSync("./setting.yaml", "utf8"));

const client = new TwitterApi({
    appKey: setting.twitter.appKey,
    appSecret: setting.twitter.appSecret,
    accessToken: setting.twitter.accessToken,
    accessSecret: setting.twitter.accessSecret,
  });

async function gettwitter() {
    const jack = client.v2.tweet("test")
}

gettwitter();