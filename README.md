# ECSEquilizer [![Build Status](https://app.travis-ci.com/b23r0/ECSEquilizer.svg?branch=main)](https://app.travis-ci.com/b23r0/ECSEquilizer)

通过阿里云OpenAPI建立一个可伸缩的负载均衡网络调度器。

# 简介

为确保代理集群网络和计算能力可以通过ECS云服务动态伸缩，所以制定实现以下策略。

节点分为static和dynamic两种，static节点是通过配置文件(`config.yaml`)实现预设的，固定不变。

dynamic由调度器根据节点集群健康状态调用ECS节点动态伸缩。

## 健康状态评估

每个节点有三个状态：great，normal，bad

评估方法用tcping，对每个节点的业务TCP端口ping四次求平均值，延迟小于100ms为great，小于1s为normal，超过1s为bad

## 负载均衡策略

调度器每10分钟调度一次，每次调度遵循三个规则。

1. 当static节点状态均大于等于normal状态（且持续超过1个小时），则释放所有dynamic节点。

2. 当所有节点当中，超过50%的节点状态为bad，则调度器自动创建N个dynamic节点，使网络状态恢复至50%节点大于等于normal。

3. 当所有节点中，超过60%的节点状态为大于等于normal，则持续释放dynamic至小于60%。

新创建的节点由于ECS初始化延迟的缘故，健康状态默认为normal，实际状态由下次调度更新实际状态。

# 编译/测试/运行

``` $> make tidy```

``` $> make test```

``` $> make ```  过程中会创建https自签证书

编辑 ```./target/config.yaml``` 修改配置信息

``` $> cd target ```

``` $> ./ecsEquilizer ```

# HTTP 身份认证

修改`config.yaml`配置文件中的`authorization`参数添加Key 

在HTTP请求头中添加 `Authorization : Bearer [vaild-key]`

# Restful API

GET /v1/nodes

```
Result:

{
	"nodes":[
		{
			"id":"S0",
			"ip":"127.0.0.1",
			"port":"1234",
			"type":"static",
			"status":"normal"
		},
		{
			"id":"S1",
			"ip":"127.0.0.1",
			"port":"4567",
			"type":"dynamic",
			"status":"normal"
		}
	]
}

```

# 回调接口

修改`config.yaml`配置文件中的`https_callback`参数添加回调地址，回调格式如下。

设置 `https_callback_auth` 参数，将在回调请求的HTTP头参数中加入`Authorization`进行身份认证

```

POST [https_callback]

{
    "id": "D1", 
    "ip": "127.0.0.1", 
    "action": "dropped"
}
```

`action` 参数值有两种，`created` 和 `dropped`，分别表示对该结点的创建和释放。

# 业务接入

业务可通过调度器接口，拿到所有节点状态，优先拿出健康状态更好的节点为客户提供服务。

调度器提供回调，当有节点创建或者被释放，将通知业务节点。

# 注意事项

修改`config.yaml`配置文件中的`work_internal_second`参数设置调度间隔时间，要确保时长足够所需调度的节点启动完成。