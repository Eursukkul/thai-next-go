'use client';

import { useEffect, useState } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';

export default function AuthCallbackPage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const code = searchParams.get('code');
    const state = searchParams.get('state');

    if (!code || !state) {
      setError('Missing code or state parameter');
      return;
    }

    // ส่ง code และ state ไปยัง backend
    fetch('http://localhost:8080/api/auth/exchange', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ code, state }),
      credentials: 'include',
    })
      .then((res) => {
        if (!res.ok) {
          return res.json().then((data) => {
            throw new Error(data.error || 'Authentication failed');
          });
        }
        return res.json();
      })
      .then(() => {
        // สำเร็จ - redirect ไป dashboard
        router.push('/dashboard');
      })
      .catch((err) => {
        setError(err.message);
      });
  }, [searchParams, router]);

  if (error) {
    return (
      <div style={{ padding: 40, textAlign: 'center' }}>
        <h1>เกิดข้อผิดพลาด</h1>
        <p style={{ color: 'red' }}>{error}</p>
        <button
          onClick={() => router.push('/')}
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
          กลับไปหน้าหลัก
        </button>
      </div>
    );
  }

  return (
    <div style={{ padding: 40, textAlign: 'center' }}>
      <h1>กำลังเข้าสู่ระบบ...</h1>
      <p>โปรดรอสักครู่</p>
    </div>
  );
}
