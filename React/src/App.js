import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Version from './Version';
import Health from './Health';

function App() {
  return (
    <BrowserRouter>
      <div className="App">
        <Routes>
          <Route path="/version" element={<Version />} />
          <Route path="/health" element={<Health />} />
        </Routes>
      </div>
    </BrowserRouter>
  );
}

export default App;
