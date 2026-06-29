const API_URL = '/api/v1';

// Check auth state on load
document.addEventListener('DOMContentLoaded', () => {
    const token = localStorage.getItem('access_token');
    const user = JSON.parse(localStorage.getItem('user') || 'null');

    if (token && user) {
        showDashboard(user);
    } else {
        showAuth();
    }
});

// UI Switching
function switchTab(tab) {
    document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
    document.querySelectorAll('.auth-form').forEach(form => form.classList.remove('active'));
    
    // Clear errors
    document.getElementById('login-error').innerText = '';
    document.getElementById('register-error').innerText = '';

    if (tab === 'login') {
        document.querySelectorAll('.tab-btn')[0].classList.add('active');
        document.getElementById('login-form').classList.add('active');
    } else {
        document.querySelectorAll('.tab-btn')[1].classList.add('active');
        document.getElementById('register-form').classList.add('active');
    }
}

function showDashboard(user) {
    document.getElementById('auth-view').classList.remove('active');
    document.getElementById('dashboard-view').classList.add('active');
    document.getElementById('user-name-display').innerText = user.name;
}

function showAuth() {
    document.getElementById('dashboard-view').classList.remove('active');
    document.getElementById('auth-view').classList.add('active');
}

// API Calls
async function handleLogin(e) {
    e.preventDefault();
    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;
    const errorEl = document.getElementById('login-error');
    const btn = e.target.querySelector('button');

    try {
        btn.innerText = 'Signing in...';
        btn.disabled = true;
        errorEl.innerText = '';

        const res = await fetch(`${API_URL}/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });

        const data = await res.json();

        if (res.ok) {
            localStorage.setItem('access_token', data.access_token);
            localStorage.setItem('user', JSON.stringify(data.user));
            showDashboard(data.user);
            e.target.reset();
        } else {
            errorEl.innerText = data.error || 'Login failed';
        }
    } catch (err) {
        errorEl.innerText = 'Network error. Please try again.';
    } finally {
        btn.innerText = 'Sign In';
        btn.disabled = false;
    }
}

async function handleRegister(e) {
    e.preventDefault();
    const name = document.getElementById('register-name').value;
    const email = document.getElementById('register-email').value;
    const password = document.getElementById('register-password').value;
    const errorEl = document.getElementById('register-error');
    const btn = e.target.querySelector('button');

    try {
        btn.innerText = 'Creating Account...';
        btn.disabled = true;
        errorEl.innerText = '';

        const res = await fetch(`${API_URL}/auth/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, email, password })
        });

        const data = await res.json();

        if (res.ok) {
            // Auto login after successful registration
            await autoLoginAfterRegister(email, password);
            e.target.reset();
        } else {
            errorEl.innerText = data.error || 'Registration failed';
        }
    } catch (err) {
        errorEl.innerText = 'Network error. Please try again.';
    } finally {
        btn.innerText = 'Create Account';
        btn.disabled = false;
    }
}

async function autoLoginAfterRegister(email, password) {
    const errorEl = document.getElementById('register-error');
    
    try {
        const res = await fetch(`${API_URL}/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });

        const data = await res.json();

        if (res.ok) {
            localStorage.setItem('access_token', data.access_token);
            localStorage.setItem('user', JSON.stringify(data.user));
            showDashboard(data.user);
        } else {
            errorEl.innerText = 'Registration successful, but auto-login failed.';
            setTimeout(() => switchTab('login'), 2000);
        }
    } catch (err) {
        errorEl.innerText = 'Registration successful. Please login.';
        setTimeout(() => switchTab('login'), 2000);
    }
}

async function handleLogout() {
    const token = localStorage.getItem('access_token');
    
    try {
        await fetch(`${API_URL}/auth/logout`, {
            method: 'POST',
            headers: { 
                'Authorization': `Bearer ${token}` 
            }
        });
    } catch(e) {
        console.error(e);
    }

    localStorage.removeItem('access_token');
    localStorage.removeItem('user');
    showAuth();
}
