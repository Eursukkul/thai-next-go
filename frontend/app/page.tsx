'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';

export default function HomePage() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // ตรวจสอบสถานะการล็อกอิน
    fetch('/api/auth/me', {
      credentials: 'include',
    })
      .then((res) => {
        if (res.ok) {
          setIsAuthenticated(true);
        }
      })
      .finally(() => setLoading(false));
  }, []);

  const handleLogin = async () => {
    const res = await fetch('/api/auth/login');
    const data = await res.json();
    if (data.auth_url) {
      window.location.href = data.auth_url;
    }
  };

  if (loading) {
    return <div style={{ padding: 40, textAlign: 'center' }}>กำลังโหลด...</div>;
  }

  if (isAuthenticated) {
    return (
      <div style={{ padding: 40, textAlign: 'center' }}>
        <h1>คุณได้ล็อกอินแล้ว</h1>
        <Link href="/dashboard">
          <button
            style={{
              padding: '12px 24px',
              fontSize: 16,
              background: '#28a745',
              color: 'white',
              border: 'none',
              borderRadius: 4,
              cursor: 'pointer',
            }}
          >
            ไปยังหน้า Dashboard
          </button>
        </Link>
      </div>
    );
  }

  return (
    <div style={{ padding: 40, textAlign: 'center' }}>
      <h1>ThaID Authentication Example</h1>
      <p>ตัวอย่างการเชื่อมต่อ ThaID ด้วย Next.js + Golang</p>
      <button
        onClick={handleLogin}
        style={{
          padding: '12px 24px',
          fontSize: 16,
          background: '#007bff',
          color: 'white',
          border: 'none',
          borderRadius: 4,
          cursor: 'pointer',
          marginTop: 20,
        }}
      >
        เข้าสู่ระบบด้วย ThaID
      </button>
    </div>
  );
}
