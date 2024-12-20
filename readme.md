一个用于监控和管理AutoDL平台中GPU的Telegram Bot程序

# 支持功能

- 监控当前GPU是否有空闲
- 启动和关闭GPU实例
- 保存和加载用户配置

# 部署步骤
1. 联系 @BotFather 创建新的 bot，并保存获得的token

2. 编译程序
   
      ```bash
      go build -o autodl-bot
      ```

3. 运行

    ```bash
    # 方式1：通过命令行参数
    ./autodl-bot --token YOUR_BOT_TOKEN
    
    # 方式2：通过环境变量
    export BOT_TOKEN=YOUR_BOT_TOKEN
    ./autodl-bot
    ```

# Bot使用方法    

- `/user xxx` 设置用户名（手机号）
- `/password xxx` 设置密码
- `/gpuvalid` 显示当前所有实例的GPU空闲情况
- `/start uuid` 启动GPU实例
- `/stop uuid` 关闭GPU实例
- `/getuser` 查看当前已设置用户

![image.png](https://s2.loli.net/2024/11/25/fJBrhIRO6zF5kZn.png)

