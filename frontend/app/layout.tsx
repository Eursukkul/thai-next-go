export const metadata = {
  title: 'ThaID Authentication - Next.js + Go',
  description: 'ThaID OAuth2 Authentication Example',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="th">
      <body style={{ margin: 0, fontFamily: 'system-ui, sans-serif' }}>
        {children}
      </body>
    </html>
  );
}
