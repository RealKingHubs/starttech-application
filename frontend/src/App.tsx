import { RouterProvider, createRouter } from '@tanstack/react-router'
import { useAuth } from '@/hooks/useAuth'
import { routeTree } from './routeTree.gen'

// Create a new router instance with the auth context
function App() {
  const auth = useAuth();
  const router = createRouter({
    routeTree,
    context: { auth },
  });
  return <RouterProvider router={router} />;
}

export default App;