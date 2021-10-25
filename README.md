# webapp-no-redirect
Web App No Redirect Hidden iFrame Sample

### To Install
``` bash
git clone https://github.com/emanor-okta/webapp-no-redirect.git
cd webapp-no-redirect/server
go mod tidy
```

### Run the Client
1. Edit **client/bin/index.js**
2. Modify `const issuer = "https://{DOMAIN}.okta.com/oauth2/default"` for your Okta Org.
3. From the **client** directory run your favorite http server to serve the content.

### Run the Server
1. Edit **server/config.yaml**
2. Modify `{CLIENT_ID}`, `{CLIENT_SECRET}`, and `{DOMAIN}` for your Okta Org and Web App client Application.
3. In Okta, for the Web App Application add `http://localhost:8082/code` as a valid **redirect URI**.
4. From the **server** directory run `go run main.go`

### Test
1. Open a browser to `http://{SERVER}:{PORT}` that your http server is running on
2. Login with a user assigned the Web App in Okta

### Extra
Because this relies on using Windows Post Message between the main window and a hidden iFrame error tracking is a bit difficult. There are three places it can happen.
1. Authn
2. Authorize
3. Token

For the **Authn** call this is taking care of by the [auth-js](https://github.com/okta/okta-auth-js) SDK running in the client application. For the **Token** call this is handled by the Web Application running in Go. For **Authorize** this can be one of two cases,
1. The error is delivered back to the redirect_uri
2. The browser redirects to the Okta hosted error page

In the case of #1, the [GO](https://golang.org/doc/install) application will handle this.    
In the case of #2 the only way to handle this is by having an [Okta Custom Domain URL](https://developer.okta.com/docs/guides/custom-url-domain/enable-the-custom-domain/) set. With this you are able to [modify the error page](https://developer.okta.com/docs/guides/custom-error-pages/edit-the-error-page/) to do a windows post message back to the parent window if one exists. If you have a Custom Domain URL this can be setup by adding the below to your custom error page javascript section.
``` javascript
      try {
        if (window.opener !== null) {
        	const e = "{{errorDescription}}".replaceAll('\'', '')
        	window.opener.postMessage('{"error":"'+e+'"}', "*");
      	}
      } catch (error) {
        console.log("error: " + error);
      }
```
