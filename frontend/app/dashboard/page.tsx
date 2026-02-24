'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';

interface UserInfo {
  sub: string;
  name: string;
  given_name: string;
  family_name: string;
  given_name_en?: string;
  family_name_en?: string;
  pid?: string;
  gender?: string;
  birthdate?: string;
  address?: string;
}

export default function DashboardPage() {
  const router = useRouter();
  const [user, setUser] = useState<UserInfo | null>(null);
  const [tokens, setTokens] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch('/api/auth/me', {
      credentials: 'include',
    })
      .then((res) => {
        if (!res.ok) {
          router.push('/');
          return null;
        }
        return res.json();
      })
      .then((data) => {
        if (data) {
          setUser(data.user);
          setTokens({
            access_token: data.access_token,
            id_token: data.id_token,
          });
        }
      })
      .finally(() => setLoading(false));
  }, [router]);

  const handleLogout = async () => {
    await fetch('/api/auth/logout', {
      method: 'GET',
      credentials: 'include',
    });
    router.push('/');
  };

  const handleIntrospect = async () => {
    if (!tokens?.access_token) return;
    
    const res = await fetch('/api/auth/introspect', {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${tokens.access_token}`,
      },
      credentials: 'include',
    });
    const data = await res.json();
    alert(JSON.stringify(data, null, 2));
  };

  if (loading) {
    return <div style={{ padding: 40 }}>กำลังโหลด...</div>;
  }

  if (!user) {
    return <div style={{ padding: 40 }}>กำลัง redirect...</div>;
  }

  return (
    <div style={{ padding: 40, maxWidth: 800, margin: '0 auto' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 30 }}>
        <h1>ข้อมูลผู้ใช้งาน</h1>
        <button
          onClick={handleLogout}
          style={{
            padding: '10px 20px',
            background: '#dc3545',
            color: 'white',
            border: 'none',
            borderRadius: 4,
            cursor: 'pointer',
          }}
        >
          ออกจากระบบ
        </button>
      </div>

      <div style={{ background: '#f8f9fa', padding: 20, borderRadius: 8, marginBottom: 20 }}>
        <h3>ข้อมูลส่วนตัว</h3>
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <tbody>
            <tr>
              <td style={{ padding: '8px 0', fontWeight: 'bold', width: 150 }}>ชื่อ-นามสกุล:</td>
              <td style={{ padding: '8px 0' }}>{user.name}</td>
            </tr>
            <tr>
              <td style={{ padding: '8px 0', fontWeight: 'bold' }}>ชื่อ (อังกฤษ):</td>
              <td style={{ padding: '8px 0' }}>
                {user.given_name_en} {user.family_name_en}
              </td>
            </tr>
            {user.pid && (
              <tr>
                <td style={{ padding: '8px 0', fontWeight: 'bold' }}>เลขประจำตัวประชาชน:</td>
                <td style={{ padding: '8px 0' }}>{user.pid}</td>
              </tr>
            )}
            {user.gender && (
              <tr>
                <td style={{ padding: '8px 0', fontWeight: 'bold' }}>เพศ:</td>
                <td style={{ padding: '8px 0' }}>{user.gender}</td>
              </tr>
            )}
            {user.birthdate && (
              <tr>
                <td style={{ padding: '8px 0', fontWeight: 'bold' }}>วันเกิด:</td>
                <td style={{ padding: '8px 0' }}>{user.birthdate}</td>
              </tr>
            )}
            {user.address && (
              <tr>
                <td style={{ padding: '8px 0', fontWeight: 'bold' }}>ที่อยู่:</td>
                <td style={{ padding: '8px 0' }}>{user.address}</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <div style={{ background: '#f8f9fa', padding: 20, borderRadius: 8 }}>
        <h3>Token Information</h3>
        <button
          onClick={handleIntrospect}
          style={{
            padding: '10px 20px',
            background: '#17a2b8',
            color: 'white',
            border: 'none',
            borderRadius: 4,
            cursor: 'pointer',
            marginBottom: 15,
          }}
        >
          ตรวจสอบ Token (Introspect)
        </button>
        
        {tokens && (
          <div>
            <h4>ID Token Claims:</h4>
            <pre
              style={{
                background: '#f4f4f4',
                padding: 15,
                borderRadius: 4,
                overflow: 'auto',
                fontSize: 12,
              }}
            >
              {JSON.stringify(user, null, 2)}
            </pre>
          </div>
        )}
      </div>
    </div>
  );
}
