<!--
 * @Author: Vincent Young
 * @Date: 2023-02-07 03:35:23
 * @LastEditors: Vincent Young
 * @LastEditTime: 2023-02-07 07:22:39
 * @FilePath: /WeiboSearcher/README.md
 * @Telegram: https://t.me/missuo
 * 
 * Copyright © 2023 by Vincent, All Rights Reserved. 
-->
# WeiboSearcher
Weibo Searcher is a query tool for the database leaked by Weibo.

## Usage
**Support any request method, including GET/POST, etc.**

## Search for mobile phone numbers by uid
```
http://localhost:11119/wb?u=1234567890
```

## Search for uid by mobile phone number
```
http://localhost:11119/wb?u=13901018888
```
## Preparing the data source
**We highly recommend you to use Clickhouse database.**

### Create table(uid)
```
CREATE TABLE wb(
uid String,
mobile String
)ENGINE = MergeTree()
    ORDER BY  (uid)
    PRIMARY KEY (uid);
```
### Create table(mobile)
```
CREATE TABLE wbm(
uid String,
mobile String
)ENGINE = MergeTree()
    ORDER BY  (mobile)
    PRIMARY KEY (mobile);
```
**We recommend that you use two tables, which can greatly speed up the efficiency of the reverse query.**

## TODO List
- [x] Support custom configuration
- [x] Support Telegram Bot
- [ ] Add frontend pages

## Statement
**We do not provide nor do we keep any leaked data.**

## Author

**WeiboSearcher** © [Vincent Young](https://github.com/missuo), Released under the [MIT](./LICENSE) License.<br>


