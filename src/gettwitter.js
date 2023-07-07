const { TwitterApi } = require('twitter-api-v2');
const yaml = require("js-yaml");
const fs = require("fs");

const setting = yaml.load(fs.readFileSync("./setting.yaml", "utf8"));
console.log(setting)

const client = new TwitterApi({
    appKey: setting.twitter.appKey,
    appSecret: setting.twitter.appSecret,
    accessToken: setting.twitter.accessToken,
    accessSecret: setting.twitter.accessSecret,
  });

async function gettwitter() {
    const homeTimeline = await client.v2.homeTimeline({ exclude: 'replies' }).then((res) => {
        console.log(res);
    }
    )
}

gettwitter();