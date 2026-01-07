namespace sz84.Core.Interfaces
{
    public interface IModuleMetadata
    {
        string Name { get; }
        string Version { get; }
        string Author { get; }
        string Description { get; }

        IEnumerable<string> RequiredModules => [];
        IEnumerable<string> OptionalModules => [];
    }

}
