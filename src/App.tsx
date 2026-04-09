import { useState } from 'react';
import { TileTemplateApp } from './components/new/TileTemplateApp';
import { ProjectsPage } from './components/new/ProjectsPage';
import { useNewTemplateStore } from './store/newTemplateStore';
import { useProjectStore } from './store/projectStore';

type Page = 'editor' | 'projects';

function App() {
  const [page, setPage] = useState<Page>('editor');

  const handleEditTemplate = (templateId: string) => {
    // Close gallery, switch to editor, load the template
    useProjectStore.getState().closeGallery();
    setPage('editor');
    useNewTemplateStore.getState().loadTemplateFromBackend(templateId);
  };

  if (page === 'projects') {
    return (
      <ProjectsPage
        onBack={() => setPage('editor')}
        onEditTemplate={handleEditTemplate}
      />
    );
  }

  return (
    <div>
      <div style={{ position: 'fixed', top: 8, right: 8, zIndex: 999 }}>
        <button
          onClick={() => setPage('projects')}
          style={{
            padding: '6px 14px',
            backgroundColor: '#6f42c1',
            color: 'white',
            border: 'none',
            borderRadius: 4,
            cursor: 'pointer',
            fontSize: 13,
            fontWeight: 500,
          }}
        >
          Projects
        </button>
      </div>
      <TileTemplateApp />
    </div>
  );
}

export default App;
