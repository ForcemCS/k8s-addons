---
name: s3_upload
description: 该 Skill 用于将本地文件上传到 RustFS 对象存储。
tools:
  - awscli
---

## RustFS 文件上传

该 Skill 用于将本地文件上传到 RustFS 对象存储。

S3 Endpoint: [http://xx.xx.0.202:9100](http://xx.xx.0.202:9100/)

Bucket: data-report

认证通过环境变量：AWS_ACCESS_KEY_ID,    AWS_SECRET_ACCESS_KEY

------

## 使用场景

当需要把文件保存到对象存储时使用，例如：

- 上传报告
- 上传日志
- 上传分析结果
- 保存生成文件

------

## 使用方法

上传文件：

```
AWS_ACCESS_KEY_ID=passwd123 \
AWS_SECRET_ACCESS_KEY=passwd123 \
aws s3 cp FILE_PATH s3://data-report/OBJECT_NAME \
--endpoint-url http://xx.xx.0.202:9100
```

------

## 示例

上传文件：

```
AWS_ACCESS_KEY_ID=passwd123 \
AWS_SECRET_ACCESS_KEY=passwd123 \
aws s3 cp /tmp/report.csv \
s3://data-report/report.csv \
--endpoint-url http://xx.xx.0.202:9100
```

上传日志：

```
AWS_ACCESS_KEY_ID=passwd123 \
AWS_SECRET_ACCESS_KEY=passwd123 \
aws s3 cp log.txt \
s3://data-report/log.txt \
--endpoint-url http://xx.xx.0.202:9100
```

------

## 说明

- 需要安装 aws-cli
- 使用 RustFS S3 API
- 认证使用环境变量
