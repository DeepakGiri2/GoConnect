const express = require('express');
const bodyParser = require('body-parser');
const cors = require('cors');

const googleApp = express();
const facebookApp = express();
const githubApp = express();

googleApp.use(cors());
googleApp.use(bodyParser.json());
googleApp.use(bodyParser.urlencoded({ extended: true }));

facebookApp.use(cors());
facebookApp.use(bodyParser.json());
facebookApp.use(bodyParser.urlencoded({ extended: true }));

githubApp.use(cors());
githubApp.use(bodyParser.json());
githubApp.use(bodyParser.urlencoded({ extended: true }));

const mockUsers = {
  google: {
    id: '108123456789012345678',
    email: 'mockuser@gmail.com',
    name: 'Mock Google User',
    picture: 'https://via.placeholder.com/150'
  },
  facebook: {
    id: '1234567890123456',
    email: 'mockuser@facebook.com',
    name: 'Mock Facebook User'
  },
  github: {
    id: 12345678,
    login: 'mockgithubuser',
    email: 'mockuser@github.com',
    name: 'Mock GitHub User'
  }
};

googleApp.get('/o/oauth2/v2/auth', (req, res) => {
  console.log('[Google] Authorization request received');
  const { redirect_uri, state, client_id } = req.query;
  const authCode = 'mock_google_auth_code_' + Date.now();
  const redirectUrl = `${redirect_uri}?code=${authCode}&state=${state}`;
  console.log(`[Google] Redirecting to: ${redirectUrl}`);
  res.redirect(redirectUrl);
});

googleApp.post('/token', (req, res) => {
  console.log('[Google] Token exchange request received');
  const { code, grant_type } = req.body;
  
  if (grant_type === 'authorization_code' && code && code.startsWith('mock_google_auth_code_')) {
    res.json({
      access_token: 'mock_google_access_token_' + Date.now(),
      refresh_token: 'mock_google_refresh_token_' + Date.now(),
      expires_in: 3600,
      token_type: 'Bearer'
    });
  } else {
    res.status(400).json({ error: 'invalid_grant' });
  }
});

googleApp.get('/oauth2/v2/userinfo', (req, res) => {
  console.log('[Google] User info request received');
  const authHeader = req.headers.authorization;
  
  if (authHeader && authHeader.startsWith('Bearer mock_google_access_token_')) {
    res.json(mockUsers.google);
  } else {
    res.status(401).json({ error: 'unauthorized' });
  }
});

facebookApp.get('/v12.0/dialog/oauth', (req, res) => {
  console.log('[Facebook] Authorization request received');
  const { redirect_uri, state } = req.query;
  const authCode = 'mock_facebook_auth_code_' + Date.now();
  const redirectUrl = `${redirect_uri}?code=${authCode}&state=${state}`;
  console.log(`[Facebook] Redirecting to: ${redirectUrl}`);
  res.redirect(redirectUrl);
});

facebookApp.get('/v12.0/oauth/access_token', (req, res) => {
  console.log('[Facebook] Token exchange request received');
  const { code } = req.query;
  
  if (code && code.startsWith('mock_facebook_auth_code_')) {
    res.json({
      access_token: 'mock_facebook_access_token_' + Date.now(),
      token_type: 'bearer',
      expires_in: 5184000
    });
  } else {
    res.status(400).json({ error: { message: 'Invalid verification code' } });
  }
});

facebookApp.get('/me', (req, res) => {
  console.log('[Facebook] User info request received');
  const { access_token } = req.query;
  
  if (access_token && access_token.startsWith('mock_facebook_access_token_')) {
    res.json(mockUsers.facebook);
  } else {
    res.status(401).json({ error: { message: 'Invalid OAuth access token' } });
  }
});

githubApp.get('/login/oauth/authorize', (req, res) => {
  console.log('[GitHub] Authorization request received');
  const { redirect_uri, state } = req.query;
  const authCode = 'mock_github_auth_code_' + Date.now();
  const redirectUrl = `${redirect_uri}?code=${authCode}&state=${state}`;
  console.log(`[GitHub] Redirecting to: ${redirectUrl}`);
  res.redirect(redirectUrl);
});

githubApp.post('/login/oauth/access_token', (req, res) => {
  console.log('[GitHub] Token exchange request received');
  const { code } = req.body;
  
  if (code && code.startsWith('mock_github_auth_code_')) {
    res.json({
      access_token: 'mock_github_access_token_' + Date.now(),
      token_type: 'bearer',
      scope: 'user:email'
    });
  } else {
    res.status(400).json({ error: 'bad_verification_code' });
  }
});

githubApp.get('/user', (req, res) => {
  console.log('[GitHub] User info request received');
  const authHeader = req.headers.authorization;
  
  if (authHeader && authHeader.startsWith('Bearer mock_github_access_token_')) {
    res.json(mockUsers.github);
  } else {
    res.status(401).json({ message: 'Requires authentication' });
  }
});

const GOOGLE_PORT = 9000;
const FACEBOOK_PORT = 9001;
const GITHUB_PORT = 9002;

googleApp.listen(GOOGLE_PORT, () => {
  console.log(`Mock Google OAuth server running on port ${GOOGLE_PORT}`);
});

facebookApp.listen(FACEBOOK_PORT, () => {
  console.log(`Mock Facebook OAuth server running on port ${FACEBOOK_PORT}`);
});

githubApp.listen(GITHUB_PORT, () => {
  console.log(`Mock GitHub OAuth server running on port ${GITHUB_PORT}`);
});
