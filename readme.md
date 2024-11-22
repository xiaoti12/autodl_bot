一个用于监控AutoDL平台中GPU空闲情况的Telegram Bot程序

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

- /user 设置用户名（手机号）

- /password 设置密码
- /gpuvalid - 显示当前所有实例的GPU空闲情况

![MWFBCN_5U_NV93___SL3FQU.png](https://s2.loli.net/2024/11/22/XP5iLljkGrFyv2J.png)

