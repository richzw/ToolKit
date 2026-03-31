- Mutter
  - 放弃老一辈审美，全力投年轻人喜欢的东西，用他们的逻辑营销
  - 资源向核心大城市集群集中 → 买医区房比学区房更重要，医院将成为最稀缺公共资源 核心城区新房或配套齐全的老房
- [境外酒店住宿](https://x.com/innomad_io/article/2030866180299538604)
  - Google Map 比价 → Agoda 不登录下单 → 立即付款
- [预测市场交易员的数学101](https://x.com/mrryanchi/article/2030564123499491329)
- [ChatGPT Won't Let You Type Until Cloudflare Reads Your React State. I Decrypted the Program That Does It](https://www.buchodi.com/chatgpt-wont-let-you-type-until-cloudflare-reads-your-react-state-i-decrypted-the-program-that-does-it/)
  - 应用层级别的机器人检测
    - 不仅仅看浏览器： 以往的机器人检测只检查浏览器环境（如 GPU、字体等），但 ChatGPT 要求你的浏览器必须已经完全加载并运行了其特定的 React 单页应用（SPA）。
    - React 状态检查： 程序会搜寻 __reactRouterContext、loaderData 和 clientBootstrap 等 React 内部数据。
  - 三层指纹采集
    - 浏览器层（Layer 1）： 检查 WebGL 渲染器、屏幕分辨率、硬件并发数、字体渲染尺寸等。
    - Cloudflare 网络层（Layer 2）： 检查地理位置、IP 地址等由 Cloudflare 边缘节点注入的头部信息，确保请求确实经过了 Cloudflare 网络。
    - 应用状态层（Layer 3）： 验证 React 组件的加载和初始化状态，这是区分高级爬虫与真人的关键。











