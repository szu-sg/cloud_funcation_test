# 模板开发说明
## 目录说明
```
-
|- conf/           
|   |- templates/                   
        |- {slug}.json(必须要)       -- 每个模板的具体配置，名字同slug
|   |- publish_conf.yaml            -- 发布区域配置       
|- images/
|   |- {slug}/                      -- 模板的图片资源存放地，名字同slug
|   |   |- icon.svg (必须要)         -- 模板的图标
|   |   |- screen_shot.jpeg(必须要)  -- 模板详情封面图
|- templates/
|   |- {slug}/                      -- 模板的代码和README资源存放地
|   |   |- README-cn.md(必须要)      -- 模板中文的README，国内详情介绍
|   |   |- README.md(必须要)         -- 模板英文README，海外详情介绍

```
## 基础概念
- 模板的Slug: 
  - 每个模板的唯一标识，以 `{project_type}_xxx` 格式为规范。
- project_type:
  - cloud_function: 云函数模板
  - native: 常规前后端模板

## 修改模板
如果要修改已有模板，请按照以下步骤：
1. 修改启动命令、插件配置等，请按照下面的**配置规范**修改`conf/templates/{slug}.json`
2. 修改代码、依赖、README等，请修改`templates/{slug}`里的内容
3. 修改封面、图标、README里的截图:
   - 截图请放在 `image/{slug}` 目录下，注意 `icon.svg` 和 `screenshot.jpeg` 是用于图标和详情封面的图片，需要联系UX同学替换。
   - 在 README.md 里以 `../../images/native_c/terminal.jpeg` 类似的方式引用image里的图片，模板发布时会自动更新到CDN。
4. 修改完成后，提交 MR 先到 **dev** 分支，dev 分支合入后会自动发布到BOE
5. BOE验证完成后，Cherry-pick对应的 MR 发布到 **master** 分支，master 分支合入后会自动发布到线上。

## 新增模板
如果要新发布一个模板，请按照以下步骤：
1. [**重要**] 联系 @wangchengbo 确定模板的运行镜像和 `slug` 
2. 在 `conf/publish_conf.yaml` 里把模板 `slug` 写在要发布的区域里
3. 其余步骤和修改模板一致。

## 配置说明
```json
{
  "BasicInfo": {
    "Name": "Node.js for AI Plugin",   // 模板显示的名字
    "Slug": "cloud_function_baas_nodejs", // slug
    "Description": "xxx",   // 描述
    "Categories": [],       // 分类，不要随意加，目前有 AI、Language(基础语言模板)、WebSite(建站模板)三种
    "ProjectType": "faas",  // 项目类型，只有faas/native两种
    "Languages": []         // 模板的主要语言，导入项目时会按此匹配Launguage类型的模板
  },
  "Runtime": {
    "RuntimeTemplate": "nodejs", // 运行的镜像，目前有 java/nodejs/python/cpp/rust/all_in_one/moonbit 7种，可以先咨询 @wangchengbo 是否满足现有需求
    "Envs": {                    // 注入个给容器环境变量，修改/创建请联系 @wangchengbo 确认
      "CLOUDIDE_CONFIG_LANGUAGE": "typescript", 
      "CLOUDIDE_CONFIG_RUNTIME": "nodejs"
    },
    "Commands": {             // 命令配置
      "BeforeStart": {        // 启动 iCube 组件进程前 hook 点，不需要更改
        "Name": ""
      },
      "AfterStart": {         // 启动 iCube 组件进程后 hook 点
        "Name": ""
      },
      "OnInit": {             // 启动完成后 hook 点，修改启动命令一般改这里
        "Name": "init",
        "Script": [           // bash -lc 固定，下面填自己的命令
          "bash",
          "-lc",
          "pnpm i"
        ]
      }
    }
  },
  "Ide": {},        // IDE的配置，插件配置，暂不开放
  "Service": {}     // 配套服务的配置，暂不开放
}
```