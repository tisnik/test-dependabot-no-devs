using System.Windows;
using System.Windows.Input;
using System.Windows.Controls;
using System.Text;
using System.Threading.Tasks;

namespace Rocnikovka
{
    public partial class MainWindow : Window
    {
        private const int MazeWidth = 20;
        private const int MazeHeight = 15;

        private int playerX = 1;
        private int playerY = 1;
        private int[,] maze;
        private int goalX;
        private int goalY;
        private bool gameEnded = false;

        // 2 = nahoru, 3 = doprava, 4 = dolů, 5 = doleva
        private int playerDirection = 2;

        public MainWindow()
        {
            InitializeComponent();
            GenerateAndDrawMaze();
            this.KeyDown += Window_KeyDown;
        }

        private void GenerateAndDrawMaze()
        {
            MazeGenerator generator = new MazeGenerator(MazeWidth, MazeHeight);
            maze = generator.Maze;

            gameEnded = false;
            playerX = 1;
            playerY = 1;
            playerDirection = 2;

            for (int y = MazeHeight - 2; y > 0 && !gameEnded; y--)
            {
                for (int x = MazeWidth - 2; x > 0; x--)
                {
                    if (maze[y, x] == 0)
                    {
                        goalX = x;
                        goalY = y;
                        break;
                    }
                }
            }

            VypisMatici();
            UpdateView();
        }

        private void VypisMatici()
        {
            int rows = maze.GetLength(0);
            int cols = maze.GetLength(1);

            int[,] vystup = (int[,])maze.Clone();
            vystup[playerY, playerX] = playerDirection;

            string vypis = "";

            for (int y = 0; y < rows; y++)
            {
                for (int x = 0; x < cols; x++)
                {
                    vypis += vystup[y, x].ToString();
                }
                vypis += "\n";
            }

            txtVystup.Text = vypis;

            VypisSubmaticiDoTextBoxu();

            UpdateView();
        }

        private void Window_KeyDown(object sender, KeyEventArgs e)
        {
            if (gameEnded)
                return;

            int dx = 0, dy = 0;

            if (e.Key == Key.NumPad4)
            {
                playerDirection = playerDirection switch
                {
                    2 => 5,
                    5 => 4,
                    4 => 3,
                    3 => 2,
                    _ => playerDirection
                };
            }
            else if (e.Key == Key.NumPad6)
            {
                playerDirection = playerDirection switch
                {
                    2 => 3,
                    3 => 4,
                    4 => 5,
                    5 => 2,
                    _ => playerDirection
                };
            }
            else if (e.Key == Key.NumPad5)
            {
                switch (playerDirection)
                {
                    case 2: dy = -1; break;
                    case 3: dx = 1; break;
                    case 4: dy = 1; break;
                    case 5: dx = -1; break;
                }
            }
            else if (e.Key == Key.NumPad2)
            {
                switch (playerDirection)
                {
                    case 2: dy = 1; break;
                    case 3: dx = -1; break;
                    case 4: dy = -1; break;
                    case 5: dx = 1; break;
                }
            }
            else if (e.Key == Key.NumPad1)
            {
                switch (playerDirection)
                {
                    case 2: dx = -1; break;
                    case 3: dy = -1; break;
                    case 4: dx = 1; break;
                    case 5: dy = 1; break;
                }
            }
            else if (e.Key == Key.NumPad3)
            {
                switch (playerDirection)
                {
                    case 2: dx = 1; break;
                    case 3: dy = 1; break;
                    case 4: dx = -1; break;
                    case 5: dy = -1; break;
                }
            }

            int newX = playerX + dx;
            int newY = playerY + dy;

            if (dx != 0 || dy != 0)
            {
                if (newX >= 0 && newX < MazeWidth && newY >= 0 && newY < MazeHeight && maze[newY, newX] == 0)
                {
                    playerX = newX;
                    playerY = newY;

                    if (playerX == goalX && playerY == goalY)
                    {
                        gameEnded = true;
                        _ = ShowFinalAnimationAsync();
                    }
                }
            }

            VypisMatici();
        }

        private async Task ShowFinalAnimationAsync()
        {
            // Skryj F1, F2, F3
            F1.Visibility = Visibility.Hidden;
            F2.Visibility = Visibility.Hidden;
            F3.Visibility = Visibility.Hidden;

            // Boční zdi
            L3.Visibility = Visibility.Visible;
            R3.Visibility = Visibility.Visible;
            L2.Visibility = Visibility.Visible;
            R2.Visibility = Visibility.Visible;
            L1.Visibility = Visibility.Visible;
            R1.Visibility = Visibility.Visible;
            L0.Visibility = Visibility.Visible;
            R0.Visibility = Visibility.Visible;

            // Level 3 ostatní
            LL3.Visibility = Visibility.Hidden;
            FL3.Visibility = Visibility.Hidden;
            FR3.Visibility = Visibility.Hidden;
            RR3.Visibility = Visibility.Hidden;

            // Level 2 ostatní
            FL2.Visibility = Visibility.Hidden;
            FR2.Visibility = Visibility.Hidden;

            // Level 1 ostatní
            FL1.Visibility = Visibility.Hidden;
            FR1.Visibility = Visibility.Hidden;

            // Konec obrázky — skryjeme
            konec1.Visibility = Visibility.Hidden;
            konec2.Visibility = Visibility.Hidden;
            konec3.Visibility = Visibility.Hidden;
            konec4.Visibility = Visibility.Hidden;

            // Animace
            await Task.Delay(500);
            konec1.Visibility = Visibility.Visible;

            await Task.Delay(500);
            konec2.Visibility = Visibility.Visible;

            await Task.Delay(500);
            konec3.Visibility = Visibility.Visible;

            await Task.Delay(500);
            konec4.Visibility = Visibility.Visible;

            await Task.Delay(500);

            MessageBox.Show("Unikl jsi z bludiště", "Konec hry", MessageBoxButton.OK, MessageBoxImage.Information);
        }

        private void NovaHra_Click(object sender, RoutedEventArgs e)
        {
            GenerateAndDrawMaze();
        }

        private void UpdateView()
        {
            // pokud je konec hry → NEUPDATEujeme view
            if (gameEnded)
                return;

            // skryjeme všechno
            string[] allNames =
            {
                "LL3","FL3","L3","F3","R3","FR3","RR3",
                "FL2","L2","F2","R2","FR2",
                "FL1","L1","F1","R1","FR1",
                "L0","R0"
            };
            foreach (var name in allNames)
                if (FindName(name) is Image img)
                    img.Visibility = Visibility.Hidden;

            // 1) view submatice
            const int viewSize = 5;
            int[,] view = new int[viewSize, viewSize];
            int rows = maze.GetLength(0), cols = maze.GetLength(1);

            int fx = 0, fy = 0, lx = 0, ly = 0;
            switch (playerDirection)
            {
                case 2: fx = 0; fy = -1; lx = -1; ly = 0; break;
                case 3: fx = 1; fy = 0; lx = 0; ly = -1; break;
                case 4: fx = 0; fy = 1; lx = 1; ly = 0; break;
                case 5: fx = -1; fy = 0; lx = 0; ly = 1; break;
            }

            for (int vy = 0; vy < viewSize; vy++)
            {
                int forwardOff = (viewSize - 1) - vy;
                for (int vx = 0; vx < viewSize; vx++)
                {
                    int leftOff = (viewSize / 2) - vx;
                    int tx = playerX + fx * forwardOff + lx * leftOff;
                    int ty = playerY + fy * forwardOff + ly * leftOff;

                    view[vy, vx] = 1;
                    if (ty >= 0 && ty < rows && tx >= 0 && tx < cols)
                        view[vy, vx] = maze[ty, tx];
                }
            }
            view[viewSize - 1, viewSize / 2] = playerDirection;

            void ShowIfWallAt(int y, int x, string imageName)
            {
                if (view[y, x] == 1 && FindName(imageName) is Image img)
                    img.Visibility = Visibility.Visible;
            }

            // Level 3
            ShowIfWallAt(1, 0, "LL3");
            ShowIfWallAt(1, 1, "FL3");
            ShowIfWallAt(1, 1, "L3");
            ShowIfWallAt(1, 2, "F3");
            ShowIfWallAt(1, 3, "R3");
            ShowIfWallAt(1, 3, "FR3");
            ShowIfWallAt(1, 4, "RR3");

            // Level 2
            ShowIfWallAt(2, 1, "FL2");
            ShowIfWallAt(2, 1, "L2");
            ShowIfWallAt(2, 2, "F2");
            ShowIfWallAt(2, 3, "R2");
            ShowIfWallAt(2, 3, "FR2");

            // Level 1
            ShowIfWallAt(3, 1, "FL1");
            ShowIfWallAt(3, 1, "L1");
            ShowIfWallAt(3, 2, "F1");
            ShowIfWallAt(3, 3, "R1");
            ShowIfWallAt(3, 3, "FR1");

            // Level 0
            ShowIfWallAt(4, 1, "L0");
            ShowIfWallAt(4, 3, "R0");
        }

        private void VypisSubmaticiDoTextBoxu()
        {
            // pokud je konec hry → nevypisujeme do submatice
            if (gameEnded)
                return;

            const int viewSize = 5;
            int[,] view = new int[viewSize, viewSize];
            int rows = maze.GetLength(0);
            int cols = maze.GetLength(1);

            int fx = 0, fy = 0, lx = 0, ly = 0;
            switch (playerDirection)
            {
                case 2: fx = 0; fy = -1; lx = -1; ly = 0; break;
                case 3: fx = 1; fy = 0; lx = 0; ly = -1; break;
                case 4: fx = 0; fy = 1; lx = 1; ly = 0; break;
                case 5: fx = -1; fy = 0; lx = 0; ly = 1; break;
            }

            for (int vy = 0; vy < viewSize; vy++)
            {
                int forwardOffset = (viewSize - 1) - vy;
                for (int vx = 0; vx < viewSize; vx++)
                {
                    int leftOffset = (viewSize / 2) - vx;
                    int targetX = playerX + fx * forwardOffset + lx * leftOffset;
                    int targetY = playerY + fy * forwardOffset + ly * leftOffset;

                    int cell = 1;
                    if (targetY >= 0 && targetY < rows && targetX >= 0 && targetX < cols)
                        cell = maze[targetY, targetX];

                    view[vy, vx] = cell;
                }
            }

            view[viewSize - 1, viewSize / 2] = playerDirection;

            var sb = new StringBuilder();
            for (int y = 0; y < viewSize; y++)
            {
                for (int x = 0; x < viewSize; x++)
                    sb.Append(view[y, x]).Append(' ');
                sb.AppendLine();
            }
            txtSubmatice.Text = sb.ToString();
        }
    }
}
 