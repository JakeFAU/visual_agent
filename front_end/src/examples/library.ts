export interface ExampleGraph {
  id: string;
  title: string;
  description: string;
  path: string;
  badges: string[];
}

const basePath = (import.meta.env.BASE_URL || '/').replace(/\/$/, '');

const withBasePath = (path: string) => `${basePath}/${path.replace(/^\//, '')}`;

export const exampleGraphs: ExampleGraph[] = [
  {
    id: 'simple-chat',
    title: 'Simple Chat',
    description: 'The smallest useful workflow: input, one LLM node, and a final output.',
    path: 'examples/simple-chat.json',
    badges: ['starter', 'text'],
  },
  {
    id: 'router',
    title: 'Structured Router',
    description: 'Classify a request into billing vs general support, then branch to the right responder.',
    path: 'examples/structured-router.json',
    badges: ['if/else', 'json'],
  },
  {
    id: 'while-loop',
    title: 'While Loop Review',
    description: 'Iterate through a review/rewrite loop with a local while-node iteration cap.',
    path: 'examples/while-loop-review.json',
    badges: ['while', 'control loop'],
  },
];

export const loadExampleGraph = async (example: ExampleGraph) => {
  const response = await fetch(withBasePath(example.path));
  if (!response.ok) {
    throw new Error(`Failed to load example '${example.title}'`);
  }

  return response.json();
};
