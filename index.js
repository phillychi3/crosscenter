const schedule = require("node-schedule");

const yaml = require("js-yaml");
const fs = require("fs");

const setting = yaml.safeLoad(fs.readFileSync("./setting.yaml", "utf8"));

//每十分中执行一次
//目標: twitter,instragram,facebook,threads
//爬取是否有新的文章
//有的話就發送到其餘的平台

const getsocial = schedule.scheduleJob(`*/${setting.waittime} * * * *`, function() {

});


