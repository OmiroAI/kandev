import importlib.util
import json
import tempfile
import unittest
from pathlib import Path


class QALockTests(unittest.TestCase):
    def test_changed_lock_is_rejected_before_qa(self):
        spec=importlib.util.spec_from_file_location('kandev_qa',Path(__file__).with_name('qa.py'))
        module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
        verify=getattr(module,'verify_locks',None)
        self.assertTrue(callable(verify),'Offline QA must verify prepared dependency identities')
        with tempfile.TemporaryDirectory() as folder:
            root=Path(folder);(root/'go.sum').write_text('changed')
            manifest=root/'locks.json';manifest.write_text(json.dumps({'go.sum':'0'*64}))
            with self.assertRaisesRegex(RuntimeError,'Dependency identity mismatch'):
                verify(root,manifest)

    def test_measured_prepared_lock_is_accepted(self):
        import hashlib
        spec=importlib.util.spec_from_file_location('kandev_qa',Path(__file__).with_name('qa.py'))
        module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
        verify=getattr(module,'verify_locks',None)
        self.assertTrue(callable(verify),'Offline QA must verify prepared dependency identities')
        with tempfile.TemporaryDirectory() as folder:
            root=Path(folder);(root/'go.sum').write_bytes(b'locked')
            manifest=root/'locks.json'
            manifest.write_text(json.dumps({'go.sum':hashlib.sha256(b'locked').hexdigest()}))
            verify(root,manifest)

    def test_added_workspace_manifest_requires_a_new_preparation_image(self):
        spec=importlib.util.spec_from_file_location('kandev_qa',Path(__file__).with_name('qa.py'))
        module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
        with tempfile.TemporaryDirectory() as folder:
            root=Path(folder);(root/'apps/new-package').mkdir(parents=True)
            (root/'apps/new-package/package.json').write_text('{}')
            manifest=root/'locks.json';manifest.write_text('{}')
            with self.assertRaisesRegex(RuntimeError,'workspace manifests'):
                module.verify_locks(root,manifest)
