import React from 'react';
import ReactDOM from 'react-dom/client';
import { ConfigProvider } from '@arco-design/web-react';
import '@arco-design/web-react/dist/css/arco.css';
import App from './App';
import { zcidTheme } from './theme/tokens';
import './styles/global.css';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ConfigProvider theme={{ primaryColor: zcidTheme.primaryOverride }}>
      <App />
    </ConfigProvider>
  </React.StrictMode>,
);
