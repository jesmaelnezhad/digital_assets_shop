export default function App() {
  return `
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Pawradise - Auth</title>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    background: #0b1020;
    color: #e6e8ef;
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    padding: 20px;
  }
  .container {
    max-width: 400px;
    width: 100%;
    padding: 40px;
    border-radius: 16px;
    background: #111827;
    border: 1px solid #1f2937;
    box-shadow: 0 20px 60px rgba(0,0,0,0.25);
  }
  h1 { margin: 0 0 8px; font-size: 28px; text-align: center; }
  .subtitle { text-align: center; margin-bottom: 24px; color: #9ca3af; font-size: 14px; }
  .form-group { margin-bottom: 16px; }
  label { display: block; margin-bottom: 6px; font-size: 14px; color: #d1d5db; }
  input {
    width: 100%;
    padding: 10px 12px;
    border-radius: 8px;
    border: 1px solid #1f2937;
    background: #0b1020;
    color: #e6e8ef;
    font-size: 14px;
  }
  input:focus { outline: none; border-color: #2563eb; }
  button {
    width: 100%;
    padding: 12px;
    border-radius: 8px;
    border: none;
    background: #2563eb;
    color: white;
    font-size: 16px;
    cursor: pointer;
    margin-top: 8px;
  }
  button:hover { background: #1d4ed8; }
  button:disabled { background: #4b5563; cursor: not-allowed; }
  .error {
    background: #7f1d1d;
    color: #fecaca;
    padding: 12px;
    border-radius: 8px;
    margin-bottom: 16px;
    font-size: 14px;
  }
  .success {
    background: #064e3b;
    color: #6ee7b7;
    padding: 12px;
    border-radius: 8px;
    margin-bottom: 16px;
    font-size: 14px;
  }
  .switch {
    text-align: center;
    margin-top: 16px;
    font-size: 14px;
    color: #9ca3af;
  }
  .switch a {
    color: #60a5fa;
    cursor: pointer;
    text-decoration: underline;
  }
  .env-badge {
    display: inline-block;
    margin-top: 18px;
    padding: 6px 12px;
    border-radius: 999px;
    font-size: 12px;
    background: #111827;
    color: #6ee7b7;
    border: 1px solid #064e3b;
  }
</style>
</head>
<body>
  <div class="container">
    <h1 id="title">Welcome</h1>
    <p class="subtitle" id="subtitle">Login to your account</p>
    <div id="message"></div>
    <form id="authForm">
      <div class="form-group" id="nameGroup" style="display:none;">
        <label for="name">Name</label>
        <input type="text" id="name" placeholder="John Doe">
      </div>
      <div class="form-group">
        <label for="email">Email</label>
        <input type="email" id="email" placeholder="you@example.com" required>
      </div>
      <div class="form-group">
        <label for="password">Password</label>
        <input type="password" id="password" placeholder="Min 6 characters" required>
      </div>
      <button type="submit" id="submitBtn">Login</button>
    </form>
    <div class="switch">
      <span id="switchText">Don't have an account? </span>
      <a id="switchLink">Register</a>
    </div>
    <div style="text-align: center; margin-top: 20px;">
      <span class="env-badge" id="env">ENV: staging</span>
    </div>
  </div>
  <script>
    const API_BASE = '/api';
    const hostname = window.location.hostname;
    document.getElementById('env').textContent = 'ENV: ' + (hostname.startsWith('staging') ? 'staging' : 'production');
    let isLogin = true;
    document.getElementById('switchLink').addEventListener('click', () => {
      isLogin = !isLogin;
      document.getElementById('title').textContent = isLogin ? 'Welcome' : 'Create Account';
      document.getElementById('subtitle').textContent = isLogin ? 'Login to your account' : 'Register a new account';
      document.getElementById('submitBtn').textContent = isLogin ? 'Login' : 'Register';
      document.getElementById('switchText').textContent = isLogin ? "Don't have an account? " : "Already have an account? ";
      document.getElementById('switchLink').textContent = isLogin ? 'Register' : 'Login';
      document.getElementById('nameGroup').style.display = isLogin ? 'none' : 'block';
      document.getElementById('message').innerHTML = '';
    });
    document.getElementById('authForm').addEventListener('submit', async (e) => {
      e.preventDefault();
      const messageDiv = document.getElementById('message');
      const submitBtn = document.getElementById('submitBtn');
      const email = document.getElementById('email').value;
      const password = document.getElementById('password').value;
      const name = document.getElementById('name').value;
      submitBtn.disabled = true;
      messageDiv.innerHTML = '';
      try {
        const endpoint = isLogin ? '/login' : '/register';
        const body = isLogin ? { email, password } : { email, password, name };
        const response = await fetch(API_BASE + endpoint, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(body)
        });
        const data = await response.json();
        if (response.ok) {
          messageDiv.innerHTML = '<div class="success">' + (isLogin ? 'Login successful!' : 'Registration successful!') + '</div>';
          if (data.token) {
            localStorage.setItem('token', data.token);
            localStorage.setItem('user', JSON.stringify(data.user));
            setTimeout(() => { alert('Welcome, ' + data.user.name + '!'); }, 500);
          }
        } else {
          messageDiv.innerHTML = '<div class="error">' + (data.error || 'Something went wrong') + '</div>';
        }
      } catch (error) {
        messageDiv.innerHTML = '<div class="error">Network error. Please try again.</div>';
      } finally {
        submitBtn.disabled = false;
      }
    });
    window.addEventListener('load', () => {
      const token = localStorage.getItem('token');
      if (token) {
        const user = JSON.parse(localStorage.getItem('user') || '{}');
        document.getElementById('subtitle').textContent = 'You are logged in';
        document.getElementById('authForm').style.display = 'none';
        document.querySelector('.switch').style.display = 'none';
        messageDiv.innerHTML = '<div class="success">Welcome back, ' + (user.name || 'User') + '!</div>';
      }
    });
  </script>
</body>
</html>
  `;
}
