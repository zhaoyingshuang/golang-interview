package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

// ============================================================
// HTTPS 与 TLS
// ============================================================

func main() {
	tlsHandshake()
	certSystem()
	goTLS()
	tlsVersions()
}

// ----------------------------------------------------------
// 1. TLS 握手流程
// ----------------------------------------------------------
func tlsHandshake() {
	fmt.Println("=== 1. TLS 握手流程 ===")
	fmt.Println()
	fmt.Println("TLS 1.2 (ECDHE 密钥交换):")
	fmt.Println("  Client                          Server")
	fmt.Println("  ─────                           ─────")
	fmt.Println("  ClientHello ──────────────────→")
	fmt.Println("  (支持的TLS版本/密码套件/随机数)")
	fmt.Println("                    ←────────── ServerHello")
	fmt.Println("                                 (选定版本/套件/随机数)")
	fmt.Println("                    ←────────── Certificate")
	fmt.Println("                                 (服务器证书链)")
	fmt.Println("                    ←────────── ServerKeyExchange")
	fmt.Println("                                 (ECDHE 参数+签名)")
	fmt.Println("                    ←────────── ServerHelloDone")
	fmt.Println("  ClientKeyExchange ────────────→")
	fmt.Println("  (ECDHE 公钥)")
	fmt.Println("  ChangeCipherSpec ─────────────→")
	fmt.Println("  Finished ────────────────────→")
	fmt.Println("                    ←────────── ChangeCipherSpec")
	fmt.Println("                    ←────────── Finished")
	fmt.Println("  ════════ 加密通信开始 ════════")
	fmt.Println()
	fmt.Println("TLS 1.3 (简化, 1-RTT):")
	fmt.Println("  ClientHello + KeyShare ───────→")
	fmt.Println("                    ←────────── ServerHello + KeyShare + Certificate + Finished")
	fmt.Println("  Finished ────────────────────→")
	fmt.Println("  ════════ 加密通信开始 ════════")
	fmt.Println()
}

// ----------------------------------------------------------
// 2. 证书体系
// ----------------------------------------------------------
func certSystem() {
	fmt.Println("=== 2. 证书体系 ===")
	fmt.Println()
	fmt.Println("证书链:")
	fmt.Println("  Root CA (自签名, 预装在浏览器/OS中)")
	fmt.Println("    └── Intermediate CA (由 Root CA 签发)")
	fmt.Println("        └── Server Certificate (由 Intermediate CA 签发)")
	fmt.Println()
	fmt.Println("验证流程:")
	fmt.Println("  1. 客户端收到服务器证书")
	fmt.Println("  2. 用签发 CA 的公钥验证签名")
	fmt.Println("  3. 沿证书链向上验证直到 Root CA")
	fmt.Println("  4. 检查证书有效期/域名匹配/吊销状态")
	fmt.Println()
	fmt.Println("证书字段:")
	fmt.Println("  Subject: CN=example.com (域名)")
	fmt.Println("  Issuer: CN=Let's Encrypt Authority X3")
	fmt.Println("  Validity: Not Before / Not After")
	fmt.Println("  Subject Alternative Name (SAN): 扩展域名列表")
	fmt.Println()
	fmt.Println("面试追问: 自签名证书和 CA 证书的区别?")
	fmt.Println("  CA 证书: 由受信任的 CA 签发, 浏览器自动信任")
	fmt.Println("  自签名: 自己签发, 客户端需手动信任 (内部服务/mTLS)")
	fmt.Println()
}

// ----------------------------------------------------------
// 3. Go TLS 编程
// ----------------------------------------------------------
func goTLS() {
	fmt.Println("=== 3. Go TLS 编程 ===")
	fmt.Println()
	fmt.Println("服务端:")
	fmt.Println("  cert, _ := tls.LoadX509KeyPair(\"server.crt\", \"server.key\")")
	fmt.Println("  config := &tls.Config{")
	fmt.Println("    Certificates: []tls.Certificate{cert},")
	fmt.Println("    MinVersion: tls.VersionTLS12,")
	fmt.Println("    CipherSuites: []uint16{")
	fmt.Println("      tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,")
	fmt.Println("      tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,")
	fmt.Println("    },")
	fmt.Println("  }")
	fmt.Println("  listener, _ := tls.Listen(\"tcp\", \":443\", config)")
	fmt.Println()
	fmt.Println("客户端:")
	fmt.Println("  config := &tls.Config{")
	fmt.Println("    InsecureSkipVerify: false, // 生产环境必须为 false")
	fmt.Println("  }")
	fmt.Println("  conn, _ := tls.Dial(\"tcp\", \"example.com:443\", config)")
	fmt.Println()
	fmt.Println("mTLS (双向认证):")
	fmt.Println("  caCert, _ := os.ReadFile(\"ca.crt\")")
	fmt.Println("  caCertPool := x509.NewCertPool()")
	fmt.Println("  caCertPool.AppendCertsFromPEM(caCert)")
	fmt.Println()
	fmt.Println("  config := &tls.Config{")
	fmt.Println("    ClientCAs:  caCertPool,             // 客户端用 CA 验证")
	fmt.Println("    ClientAuth: tls.RequireAndVerifyClientCert, // 要求客户端证书")
	fmt.Println("    Certificates: []tls.Certificate{serverCert},")
	fmt.Println("  }")
	fmt.Println()
}

// ----------------------------------------------------------
// 4. TLS 版本对比
// ----------------------------------------------------------
func tlsVersions() {
	fmt.Println("=== 4. TLS 版本对比 ===")
	fmt.Println()
	fmt.Println("  ┌──────────┬────────────┬──────────────────────────┐")
	fmt.Println("  │ 版本     │ 握手 RTT   │ 安全性                   │")
	fmt.Println("  ├──────────┼────────────┼──────────────────────────┤")
	fmt.Println("  │ SSL 3.0  │ 2-RTT      │ 不安全, 已弃用           │")
	fmt.Println("  │ TLS 1.0  │ 2-RTT      │ 不安全, 已弃用 (POODLE)  │")
	fmt.Println("  │ TLS 1.1  │ 2-RTT      │ 不安全, 已弃用 (BEAST)   │")
	fmt.Println("  │ TLS 1.2  │ 2-RTT      │ 安全, 当前主流           │")
	fmt.Println("  │ TLS 1.3  │ 1-RTT      │ 最安全, 0-RTT 可选       │")
	fmt.Println("  └──────────┴────────────┴──────────────────────────┘")
	fmt.Println()
	fmt.Println("  TLS 1.3 改进:")
	fmt.Println("    1. 握手从 2-RTT 缩短到 1-RTT (0-RTT 恢复)")
	fmt.Println("    2. 移除了不安全的密码套件 (RC4/3DES/CBC)")
	fmt.Println("    3. 所有握手消息都加密 (TLS 1.2 的部分是明文)")
	fmt.Println("    4. 只支持 ECDHE 密钥交换 (前向保密)")
	fmt.Println()
	fmt.Println("  面试追问: HTTPS 的性能开销?")
	fmt.Println("    TLS 握手: 1-2 次 RTT (TLS 1.3 只需 1 次)")
	fmt.Println("    加解密: AES-GCM 硬件加速, 开销很小")
	fmt.Println("    证书验证: 证书链验证 + OCSP/CRL 检查")
	fmt.Println("    优化: Session Resumption / TLS 1.3 0-RTT")
	fmt.Println()
}

// 生成自签名证书的辅助函数 (演示用)
func generateSelfSignedCert() {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{CommonName: "localhost"},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}
	certDER, _ := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	_ = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
}
