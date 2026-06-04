---
title: HTTPS 与 TLS
---

## 1. TLS 握手流程

**TLS 1.2 (ECDHE, 2-RTT)**：
1. ClientHello（支持的版本/密码套件/随机数）
2. ServerHello + Certificate + ServerKeyExchange（ECDHE 参数+签名）
3. ClientKeyExchange（ECDHE 公钥）→ 双方计算出预主密钥
4. ChangeCipherSpec + Finished（双方）

**TLS 1.3 (1-RTT)**：ClientHello 直接带上 KeyShare，Server 一次回完所有内容。

::: tip TLS 1.3 改进
1. 握手从 2-RTT 缩短到 1-RTT（0-RTT 恢复可选）
2. 移除不安全密码套件（RC4/3DES/CBC）
3. 所有握手消息加密（TLS 1.2 部分明文）
4. 只支持 ECDHE（前向保密）
:::

## 2. 加密体系

- **对称加密**：AES-GCM（数据加密，快）
- **非对称加密**：RSA/ECDSA（密钥交换/签名，慢）
- **混合加密**：非对称交换密钥 → 对称加密数据

## 3. 证书体系

```
Root CA (自签名, 预装在浏览器/OS)
└── Intermediate CA (由 Root CA 签发)
    └── Server Certificate (由 Intermediate CA 签发, 包含域名)
```

验证：收到证书 → 沿证书链向上验证签名 → 检查有效期/域名/吊销状态。

## 4. Go TLS 编程

```go
// 服务端
cert, _ := tls.LoadX509KeyPair("server.crt", "server.key")
config := &tls.Config{
    Certificates: []tls.Certificate{cert},
    MinVersion:   tls.VersionTLS12,
}
listener, _ := tls.Listen("tcp", ":443", config)

// mTLS (双向认证)
caCert, _ := os.ReadFile("ca.crt")
caCertPool := x509.NewCertPool()
caCertPool.AppendCertsFromPEM(caCert)
config := &tls.Config{
    ClientCAs:    caCertPool,
    ClientAuth:   tls.RequireAndVerifyClientCert,
    Certificates: []tls.Certificate{serverCert},
}
```

::: warning 面试追问：HTTPS 性能开销
TLS 握手 1-2 RTT（TLS 1.3 只需 1 次），加解密有 AES-NI 硬件加速开销很小。优化：Session Resumption / TLS 1.3 0-RTT。
:::
