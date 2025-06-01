package examples

import (
	"context"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"log"
	"net/http"
)

func getToken() (string, error) {
	const (
		realm         = "sv"
		clientSecret  = "J8aeMmU0sJfSvaZCQGsVkQCuYMk6NlnT"
		authServerURL = "https://sso.russionbear.com"
		clientID      = "sv-mr" // 替换为你的 clientID
	)
	username := "russionbear@163.com"
	password := "123456"

	// 配置 OAuth2 客户端
	oauth2Config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/realms/sv/protocol/openid-connect/auth", authServerURL),
			TokenURL: fmt.Sprintf("%s/realms/sv/protocol/openid-connect/token", authServerURL),
		},
	}

	// 使用用户名和密码获取 token
	token, err := oauth2Config.PasswordCredentialsToken(context.Background(), username, password)
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

func getUserInfo(token *oauth2.Token, ctx context.Context) {
	clientID := "sv-mr"
	clientSecret := "J8aeMmU0sJfSvaZCQGsVkQCuYMk6NlnT"
	realm := "sv"
	authServerURL := "https://sso.russionbear.com/"

	// 创建 OIDC 提供者实例
	provider, err := oidc.NewProvider(ctx, authServerURL+"realms/"+realm)
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	// 配置 OAuth2 客户端
	var oauth2Config = oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  authServerURL + "realms/" + realm + "/protocol/openid-connect/auth",
			TokenURL: authServerURL + "realms/" + realm + "/protocol/openid-connect/token",
		},
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		_ = fmt.Errorf("没有找到 ID Token")
		return
	}
	idToken, err := provider.Verifier(&oidc.Config{ClientID: clientID}).Verify(ctx, rawIDToken)
	if err != nil {
		_ = fmt.Sprintf("验证 ID Token 失败: %v", err)
		return
	}
	var claims struct {
		Sub               string `json:"sub"`
		Name              string `json:"name"`
		GivenName         string `json:"given_name"`
		FamilyName        string `json:"family_name"`
		Email             string `json:"email"`
		EmailVerified     bool   `json:"email_verified"`
		Picture           string `json:"picture"`
		PreferredUsername string `json:"preferred_username"`
		Profile           string `json:"profile"`
		ZoneInfo          string `json:"zoneinfo"`
		Locale            string `json:"locale"`
		UpdatedAt         int64  `json:"updated_at"`
		Exp               int64  `json:"exp"`
	}
	if err := idToken.Claims(&claims); err != nil {
		fmt.Sprintf("解析声明失败: %v", err)
		return
	}
	fmt.Printf("欢迎, %s!", claims.Name)

	// 使用刷新令牌获取新的访问令牌
	newToken, err := oauth2Config.TokenSource(ctx, token).Token()
	if err != nil {
		_ = fmt.Sprintf("刷新令牌失败: %v", err)
		return
	}
	fmt.Println("新的访问令牌:", newToken.AccessToken)
}

func refreshToken() {
	const (
		realm         = "sv"
		clientSecret  = "J8aeMmU0sJfSvaZCQGsVkQCuYMk6NlnT"
		authServerURL = "https://sso.russionbear.com"
		clientID      = "sv-mr" // 替换为你的 clientID
	)
	username := "russionbear@163.com"
	password := "123456"

	// 配置 OAuth2 客户端
	oauth2Config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/realms/sv/protocol/openid-connect/auth", authServerURL),
			TokenURL: fmt.Sprintf("%s/realms/sv/protocol/openid-connect/token", authServerURL),
		},
	}

	// 使用用户名和密码获取 token
	token, err := oauth2Config.PasswordCredentialsToken(context.Background(), username, password)
	if err != nil {
		_ = fmt.Errorf("获取 token 失败: %v", err)
	}

	fmt.Println("访问令牌:", token.AccessToken)

	newToken, err := oauth2Config.TokenSource(context.Background(), token).Token()
	if err != nil {
		fmt.Sprintf("刷新令牌失败: %v", err)
		return
	}
	fmt.Println("新的访问令牌:", newToken.AccessToken)
}

func main() {
	//token := lib.GetToken("1954586261@qq.com", "1234")
	//fmt.Println(token)
	//if len(os.Args) < 2 {
	//	fmt.Println("Usage: go run main.go server_url|client")
	//	os.Exit(1)
	//}
	//if os.Args[1] == "client" {
	//	if len(os.Args) < 4 {
	//		fmt.Println("Usage: go run main.go server_url|client username password")
	//		os.Exit(1)
	//	}
	//	client()
	//} else {
	//	server(os.Args[1])
	//}

	// Keycloak 配置信息
	clientID := "sv-mr"
	clientSecret := "J8aeMmU0sJfSvaZCQGsVkQCuYMk6NlnT"
	realm := "sv"
	authServerURL := "https://sso.russionbear.com/"

	// 创建 OIDC 提供者实例
	provider, err := oidc.NewProvider(context.Background(), authServerURL+"realms/"+realm)
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	// 配置 OAuth2 客户端
	oauth2Config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8080/callback", // 需要手动配置
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		Endpoint:     provider.Endpoint(),
	}

	// 生成认证 URL
	state := "random-state-string" // 在实际应用中应使用安全随机字符串
	url := oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	fmt.Println("请访问以下 URL 进行登录、注册和授权:", url)

	// 启动 HTTP 服务器处理回调
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "获取授权码失败", http.StatusBadRequest)
			return
		}
		token, err := oauth2Config.Exchange(r.Context(), code)
		if err != nil {
			http.Error(w, fmt.Sprintf("交换令牌失败: %v", err), http.StatusInternalServerError)
			return
		}
		rawIDToken, ok := token.Extra("id_token").(string)
		if !ok {
			http.Error(w, "没有找到 ID Token", http.StatusInternalServerError)
			return
		}
		idToken, err := provider.Verifier(&oidc.Config{ClientID: clientID}).Verify(r.Context(), rawIDToken)
		if err != nil {
			http.Error(w, fmt.Sprintf("验证 ID Token 失败: %v", err), http.StatusInternalServerError)
			return
		}
		var claims struct {
			Sub               string `json:"sub"`
			Name              string `json:"name"`
			GivenName         string `json:"given_name"`
			FamilyName        string `json:"family_name"`
			Email             string `json:"email"`
			EmailVerified     bool   `json:"email_verified"`
			Picture           string `json:"picture"`
			PreferredUsername string `json:"preferred_username"`
			Profile           string `json:"profile"`
			ZoneInfo          string `json:"zoneinfo"`
			Locale            string `json:"locale"`
			UpdatedAt         int64  `json:"updated_at"`
			Exp               int64  `json:"exp"`
		}
		if err := idToken.Claims(&claims); err != nil {
			http.Error(w, fmt.Sprintf("解析声明失败: %v", err), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "欢迎, %s!", claims.Name)

		// 使用刷新令牌获取新的访问令牌
		newToken, err := oauth2Config.TokenSource(r.Context(), token).Token()
		if err != nil {
			http.Error(w, fmt.Sprintf("刷新令牌失败: %v", err), http.StatusInternalServerError)
			return
		}
		fmt.Println("新的访问令牌:", newToken.AccessToken)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))

}
