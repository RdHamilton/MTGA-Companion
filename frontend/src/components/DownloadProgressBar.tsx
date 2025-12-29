import { useDownload } from '@/context/DownloadContext';
import './DownloadProgressBar.css';

const DownloadProgressBar = () => {
  const { state, isDownloading } = useDownload();

  // Don't render if nothing is downloading
  if (!isDownloading && state.tasks.length === 0) {
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
