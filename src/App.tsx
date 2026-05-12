import { TileTemplateApp } from './components/new/TileTemplateApp';
import { TemplateSidebar } from './components/new/TemplateSidebar';

function App() {
  return (
    <div style={{ display: 'flex', height: '100vh', overflow: 'hidden' }}>
      <TemplateSidebar />
      <main style={{ flex: 1, overflow: 'auto' }}>
        <TileTemplateApp />
      </main>
    </div>
  );
}

export default App;
