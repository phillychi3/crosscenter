const request = require("request");
const yaml = require("js-yaml");
const fs = require("fs");
const setting = yaml.load(fs.readFileSync("./setting.yaml", "utf8"));

module.exports = function gettwitter3(name) {
  const options = {
    method: "GET",
    url: "https://twttrapi.p.rapidapi.com/get-tweet",
    qs: {
      username: name,
    },
    headers: {
      "X-RapidAPI-Key": setting.twitter.rapid,
      "X-RapidAPI-Host": "twttrapi.p.rapidapi.com",
    },
  };

  request(options, function (error, response, body) {
    if (error) throw new Error(error);
    return body;
  });
};
