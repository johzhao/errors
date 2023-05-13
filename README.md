# errors

## 目的

解决后端服务如何返回错误码的问题
  * 任意函数可以返回特定的错误码
  * 不返回特定错误码的函数只需要简单使用 fmt.Errorf 来 wrap 一下返回的错误并返回
  * 最终返回给用户前可以从 error 对象中找到最初的 ResponseError 对象,并生成错误信息返回

## 使用方法

* 定义服务自己的 BusinessError 实例
* 在需要返回特定错误的地方使用 NewResponseError 创建 ResponseError 对象并返回
  ```go
  return NewResponseError(ErrorInvalidParameter, errors.New("invalid parameter"))
  ```
* 在 controller/handler 里,如果 service 返回错误的话,创建一个 ResponseError 对象并返回,这样返回该 API 一个默认的错误
  ```go
  return NewResponseError(ErrorCreateUser, errors.New("failed to create user"))
  ```
* 在 web 框架的错误处理 handler 里调用 ConvertToResponseError 通过 error 对象获取到最初的 ResponseError 对象,拿到需要返回给客户的错误码以及错误提示
  ```go
  responseErr := ConvertToResponseError(err, ErrorUnknown)
  ctx.Abort()
  ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
    "code": responseErr.Code,
    "message": responseErr.Message,
  })
  ```

## 示例代码

见测试代码
