import { useDownload } from '@/context/DownloadContext';
import './DownloadProgressBar.css';

const DownloadProgressBar = () => {
  const { state } = useDownload();

  // Don't render if no active task to display
  if (!state.activeTask) {
    return null;
  }

  const activeTask = state.activeTask;

  return (
    <div className="download-progress-container">
      {activeTask && (
        <>
          <div className="download-progress-bar">
            <div
              className={`download-progress-fill ${activeTask.status}`}
              style={{ width: `${activeTask.progress}%` }}
            />
          </div>
          <span className="download-progress-text">
            {activeTask.status === 'error' ? (
              <span className="download-error">{activeTask.error || 'Download failed'}</span>
            ) : (
              <>
                {activeTask.description}
                {activeTask.progress > 0 && activeTask.progress < 100 && (
                  <span className="download-percentage"> ({Math.round(activeTask.progress)}%)</span>
                )}
              </>
            )}
          </span>
        </>
      )}
    </div>
  );
};

export default DownloadProgressBar;
