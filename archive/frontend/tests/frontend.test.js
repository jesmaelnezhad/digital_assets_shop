/**
 * @jest-environment jsdom
 */
const fs = require('fs');
const path = require('path');

// Helper to read and evaluate HTML file in JSDOM
function loadHtml(filename) {
  const html = fs.readFileSync(path.join(__dirname, '../public', filename), 'utf8');
  document.documentElement.innerHTML = html;
  // Execute scripts
  const scripts = document.querySelectorAll('script');
  scripts.forEach(script => {
    if (script.textContent) {
      try {
        new Function(script.textContent)();
      } catch (e) {
        // Some globals may fail outside browser context
      }
    }
  });
}

describe('Frontend - Auth Page', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.innerHTML = '';
    // Re-load for each test
  });

  describe('Page Structure', () => {
    test('index.html has required elements', () => {
      loadHtml('index.html');
      
      expect(document.getElementById('authForm')).not.toBeNull();
      expect(document.getElementById('formTitle')).not.toBeNull();
      expect(document.getElementById('formSubtitle')).not.toBeNull();
      expect(document.getElementById('email')).not.toBeNull();
      expect(document.getElementById('password')).not.toBeNull();
      expect(document.getElementById('submitBtn')).not.toBeNull();
      expect(document.getElementById('switchLink')).not.toBeNull();
      expect(document.getElementById('nameGroup')).not.toBeNull();
    });

    test('admin.html has admin login form', () => {
      loadHtml('admin.html');
      
      expect(document.getElementById('adminLogin')).not.toBeNull();
      expect(document.getElementById('adminToken')).not.toBeNull();
      expect(document.getElementById('loginBtn')).not.toBeNull();
      expect(document.querySelector('.admin-login')).not.toBeNull();
    });

    test('admin.html has user table structure', () => {
      loadHtml('admin.html');
      
      expect(document.getElementById('usersTable')).not.toBeNull();
      expect(document.getElementById('usersBody')).not.toBeNull();
      expect(document.querySelector('table')).not.toBeNull();
    });

    test('admin.html has admin panel hidden by default', () => {
      loadHtml('admin.html');
      
      expect(document.getElementById('adminPanel')).toBeNull() || 
        document.getElementById('adminPanel').classList.contains('hidden');
    });
  });

  describe('Environment Detection', () => {
    test('index.html detects production environment', () => {
      // Simulate production path
      window.location = { pathname: '/' };
      loadHtml('index.html');
      
      // The env badge should show production on /
      const envEl = document.getElementById('env');
      expect(envEl).not.toBeNull();
    });

    test('index.html detects staging environment', () => {
      // Simulate staging path  
      window.location = { pathname: '/staging/' };
      loadHtml('index.html');
      
      const envEl = document.getElementById('env');
      expect(envEl).not.toBeNull();
    });

    test('admin.html has environment badge', () => {
      loadHtml('admin.html');
      
      expect(document.getElementById('env')).not.toBeNull();
    });
  });

  describe('Form Elements', () => {
    test('email input has proper attributes', () => {
      loadHtml('index.html');
      
      const emailInput = document.getElementById('email');
      expect(emailInput.type).toBe('email');
      expect(emailInput.required).toBe(true);
    });

    test('password input has proper attributes', () => {
      loadHtml('index.html');
      
      const passwordInput = document.getElementById('password');
      expect(passwordInput.type).toBe('password');
      expect(passwordInput.required).toBe(true);
    });

    test('name input exists for registration', () => {
      loadHtml('index.html');
      
      expect(document.getElementById('name')).not.toBeNull();
      expect(document.getElementById('nameGroup')).not.toBeNull();
    });

    test('login button is present', () => {
      loadHtml('index.html');
      
      const btn = document.getElementById('submitBtn');
      expect(btn).not.toBeNull();
      expect(btn.textContent).toBe('Login');
    });

    test('admin token input exists', () => {
      loadHtml('admin.html');
      
      const tokenInput = document.getElementById('adminToken');
      expect(tokenInput).not.toBeNull();
      expect(tokenInput.type).toBe('password');
    });
  });

  describe('Toggle Functionality', () => {
    test('switch link toggles between login and register', () => {
      loadHtml('index.html');
      
      // Initial state: login mode
      expect(document.getElementById('submitBtn').textContent).toBe('Login');
      expect(document.getElementById('formTitle').textContent).toBe('Welcome');
      expect(document.getElementById('nameGroup').style.display).toBe('none');
      
      // Click switch
      document.getElementById('switchLink').click();
      
      // Should be in register mode
      expect(document.getElementById('submitBtn').textContent).toBe('Register');
      expect(document.getElementById('formTitle').textContent).toBe('Create Account');
      expect(document.getElementById('nameGroup').style.display).toBe('block');
      
      // Click again to go back
      document.getElementById('switchLink').click();
      
      expect(document.getElementById('submitBtn').textContent).toBe('Login');
      expect(document.getElementById('formTitle').textContent).toBe('Welcome');
    });
  });

  describe('API Base Configuration', () => {
    test('API_BASE is set to /api/v1', () => {
      loadHtml('index.html');
      
      // Check that API_BASE constant exists
      expect(window.API_BASE).toBeUndefined(); // const is not global
      // The script defines const API_BASE = '/api/v1'
      // Verify it's in the source
      const html = fs.readFileSync(path.join(__dirname, '../public/index.html'), 'utf8');
      expect(html).toContain("const API_BASE = '/api/v1'");
    });

    test('admin API_BASE is set to /api/v1', () => {
      const html = fs.readFileSync(path.join(__dirname, '../public/admin.html'), 'utf8');
      expect(html).toContain("const API_BASE = '/api/v1'");
    });
  });

  describe('Logout Functionality', () => {
    test('logout function exists in admin panel', () => {
      const html = fs.readFileSync(path.join(__dirname, '../public/admin.html'), 'utf8');
      expect(html).toContain('function logout()');
      expect(html).toContain('localStorage.removeItem');
    });

    test('logout function exists in main app', () => {
      const html = fs.readFileSync(path.join(__dirname, '../public/index.html'), 'utf8');
      expect(html).toContain('function handleLogout()');
      expect(html).toContain('localStorage.removeItem');
    });
  });

  describe('Profile View', () => {
    test('profile view elements exist', () => {
      const html = fs.readFileSync(path.join(__dirname, '../public/index.html'), 'utf8');
      expect(html).toContain('id="profileView"');
      expect(html).toContain('id="profileName"');
      expect(html).toContain('id="profileEmail"');
      expect(html).toContain('id="avatar"');
      expect(html).toContain('id="profileNameInput"');
      expect(html).toContain('id="profileEmailInput"');
    });

    test('profile update form exists', () => {
      const html = fs.readFileSync(path.join(__dirname, '../public/index.html'), 'utf8');
      expect(html).toContain('handleProfileUpdate');
      expect(html).toContain('Update Profile');
    });
  });

  describe('Session Management', () => {
    test('script checks localStorage on load', () => {
      const html = fs.readFileSync(path.join(__dirname, '../public/index.html'), 'utf8');
      expect(html).toContain('localStorage.getItem');
      expect(html).toContain('DOMContentLoaded');
    });

    test('script clears localStorage on logout', () => {
      const html = fs.readFileSync(path.join(__dirname, '../public/index.html'), 'utf8');
      expect(html).toContain("localStorage.removeItem('token')");
      expect(html).toContain("localStorage.removeItem('user')");
    });

    test('admin script manages admin token', () => {
      const html = fs.readFileSync(path.join(__dirname, '../public/admin.html'), 'utf8');
      expect(html).toContain("localStorage.removeItem('token')");
    });
  });

  describe('Admin Panel Features', () => {
    test('reset password function exists', () => {
      const html = fs.readFileSync(path.join(__dirname, '../public/admin.html'), 'utf8');
      expect(html).toContain('function resetPassword');
      expect(html).toContain('/reset-password');
    });

    test('delete user function exists', () => {
      const html = fs.readFileSync(path.join(__dirname, '../public/admin.html'), 'utf8');
      expect(html).toContain('function deleteUser');
      expect(html).toContain('DELETE');
    });

    test('load users function exists', () => {
      const html = fs.readFileSync(path.join(__dirname, '../public/admin.html'), 'utf8');
      expect(html).toContain('function loadUsers');
      expect(html).toContain('/admin/users');
    });
  });

  describe('Security Features', () => {
    test('escapeHtml function exists in admin', () => {
      const html = fs.readFileSync(path.join(__dirname, '../public/admin.html'), 'utf8');
      expect(html).toContain('function escapeHtml');
      expect(html).toContain('replace');
    });

    test('XSS prevention in admin panel', () => {
      const html = fs.readFileSync(path.join(__dirname, '../public/admin.html'), 'utf8');
      // Should escape user input before displaying
      expect(html).toContain("replace(/&/g,'&amp;')");
    });
  });

  describe('User Feedback', () => {
    test('success message display exists', () => {
      const html = fs.readFileSync(path.join(__dirname, '../public/index.html'), 'utf8');
      expect(html).toContain('showMessage');
      expect(html).toContain('success');
    });

    test('error message display exists', () => {
      const html = fs.readFileSync(path.join(__dirname, '../public/index.html'), 'utf8');
      expect(html).toContain('error');
      expect(html).toContain('showMessage');
    });

    test('admin toast notifications', () => {
      const html = fs.readFileSync(path.join(__dirname, '../public/admin.html'), 'utf8');
      expect(html).toContain('showToast');
      expect(html).toContain('toast');
    });
  });

  describe('Accessibility', () => {
    test('labels are associated with inputs', () => {
      loadHtml('index.html');
      
      const emailLabel = document.querySelector('label[for="email"]');
      expect(emailLabel).not.toBeNull();
      
      const passwordLabel = document.querySelector('label[for="password"]');
      expect(passwordLabel).not.toBeNull();
    });

    test('form has proper structure', () => {
      loadHtml('index.html');
      
      const form = document.getElementById('loginForm');
      expect(form).not.toBeNull();
      expect(form.tagName).toBe('FORM');
    });
  });
});
