using System;

namespace Rocnikovka
{
    public class MazeGenerator
    {
        private readonly int width;
        private readonly int height;
        private readonly int[,] maze;
        private readonly Random rand = new Random();

        public int[,] Maze => maze;

        public MazeGenerator(int width, int height)
        {
            this.width = width % 2 == 0 ? width - 1 : width;
            this.height = height % 2 == 0 ? height - 1 : height;
            maze = new int[this.height, this.width];
            Generate();
        }

        private void Generate()
        {
            for (int y = 0; y < height; y++)
                for (int x = 0; x < width; x++)
                    maze[y, x] = 1;

            int cx = 1;
            int cy = 1;
            maze[cy, cx] = 0;

            while (true)
            {
                var dirs = GetShuffledDirections();
                bool moved = false;

                foreach (var (dx, dy) in dirs)
                {
                    int nx = cx + dx * 2;
                    int ny = cy + dy * 2;

                    if (IsInBounds(nx, ny) && maze[ny, nx] == 1)
                    {
                        maze[cy + dy, cx + dx] = 0;
                        maze[ny, nx] = 0;
                        cx = nx;
                        cy = ny;
                        moved = true;
                        break;
                    }
                }

                if (moved)
                    continue;

                bool found = false;
                for (int y = 1; y < height; y += 2)
                {
                    for (int x = 1; x < width; x += 2)
                    {
                        if (maze[y, x] == 1 && HasVisitedNeighbor(x, y))
                        {
                            maze[y, x] = 0;
                            var neighbors = GetVisitedNeighbors(x, y);
                            var (dx, dy) = neighbors[rand.Next(neighbors.Length)];
                            maze[y + dy / 2, x + dx / 2] = 0;
                            cx = x;
                            cy = y;
                            found = true;
                            break;
                        }
                    }
                    if (found) break;
                }

                if (!found)
                    break;
            }
        }

        private (int dx, int dy)[] GetShuffledDirections()
        {
            var dirs = new (int dx, int dy)[] { (0, -1), (1, 0), (0, 1), (-1, 0) };
            for (int i = dirs.Length - 1; i > 0; i--)
            {
                int j = rand.Next(i + 1);
                (dirs[i], dirs[j]) = (dirs[j], dirs[i]);
            }
            return dirs;
        }

        private bool IsInBounds(int x, int y) =>
            x > 0 && y > 0 && x < width && y < height;

        private bool HasVisitedNeighbor(int x, int y)
        {
            foreach (var (dx, dy) in GetShuffledDirections())
            {
                int nx = x + dx * 2;
                int ny = y + dy * 2;
                if (IsInBounds(nx, ny) && maze[ny, nx] == 0)
                    return true;
            }
            return false;
        }

        private (int dx, int dy)[] GetVisitedNeighbors(int x, int y)
        {
            var list = new System.Collections.Generic.List<(int dx, int dy)>();
            foreach (var (dx, dy) in GetShuffledDirections())
            {
                int nx = x + dx * 2;
                int ny = y + dy * 2;
                if (IsInBounds(nx, ny) && maze[ny, nx] == 0)
                    list.Add((dx * 2, dy * 2));
            }
            return list.ToArray();
        }
    }
}
