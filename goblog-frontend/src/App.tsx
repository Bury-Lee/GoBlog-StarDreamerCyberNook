import { BrowserRouter, Routes, Route } from 'react-router-dom';
import HomePage from './pages/Home';
import LoginPage from './pages/User/Login';
import RegisterPage from './pages/User/Register';
import ArticleDetailPage from './pages/Article/Detail';
import ArticleListPage from './pages/Article/List';
import ProfilePage from './pages/User/Profile';
import EditorPage from './pages/Article/Editor';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/articles" element={<ArticleListPage />} />
        <Route path="/article/:id" element={<ArticleDetailPage />} />
        <Route path="/editor" element={<EditorPage />} />
        <Route path="/editor/:id" element={<EditorPage />} />
        <Route path="/profile" element={<ProfilePage />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
