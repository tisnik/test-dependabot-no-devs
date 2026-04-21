"""Benchmarks for app.database module."""

from typing import Optional
from pathlib import Path
import pytest
import random
from sqlalchemy.orm import Session
from pytest_benchmark.fixture import BenchmarkFixture
from datetime import UTC, datetime

from app import database
from app.database import get_session
from configuration import configuration
from utils.suid import get_suid
from models.database.conversations import UserConversation

# number of records to be stored in database before benchmarks
MIDDLE_DB_RECORDS_COUNT = 1000
LARGE_DB_RECORDS_COUNT = 10000


@pytest.fixture(name="configuration_filename")
def configuration_filename_fixture() -> str:
    """Retrieve configuration file name to be used by integration tests."""
    return "tests/configuration/benchmarks-lightspeed-stack.yaml"


@pytest.fixture(name="sqlite_database")
def sqlite_database_fixture(configuration_filename: str, tmp_path: Path) -> None:
    """Initialize a temporary SQLite database for benchmarking."""
    # try to load the configuration containing SQLite database setup
    configuration.load_configuration(configuration_filename)
    assert configuration.database_configuration.sqlite is not None

    # we need to start each benchmark with empty database
    configuration.database_configuration.sqlite.db_path = str(tmp_path / "database.db")

    # initialize database session and create tables
    database.initialize_database()
    database.create_tables()


def generate_provider() -> str:
    """Generate provider name."""
    providers = [
        "openai",
        "azure",
        "vertexAI",
        "watsonx",
        "RHOAI (vLLM)",
        "RHAIIS (vLLM)",
        "RHEL AI (vLLM)",
    ]
    return random.choice(providers)


def generate_model_for_provider(provider: str) -> str:
    """Generate model ID for given provider."""
    models: dict[str, list[str]] = {
        "openai": [
            "gpt-5",
            "gpt-5.2",
            "gpt-5.2 pro",
            "gpt-5 mini",
            "gpt-4.1",
            "gpt-4o",
            "gpt-4-turbo",
            "gpt-4.1 mini",
            "gpt-4.1 nano",
            "o4-mini",
            "o1",
            "o3",
            "o4",
        ],
        "azure": [
            "gpt-5",
            "gpt-5-mini",
            "gpt-5-nano",
            "gpt-4.1",
            "gpt-4.1 mini",
            "gpt-5-chat",
            "gpt-5.1",
            "gpt-5.1-codex",
            "gpt-5.2",
            "gpt-5.2-chat",
            "gpt-5.2-codex",
            "claude-opus-4-5",
            "claude-haiku-4-5",
            "claude-sonnet-4-5",
            "DeepSeek-v3.1",
        ],
        "vertexAI": [
            "google/gemini-2.0-flash",
            "google/gemini-2.5-flash",
            "google/gemini-2.5-pro",
        ],
        "watsonx": [
            "all-mini-l6-v2",
            "multilingual-e5-large",
            "granite-embedding-107m-multilingual",
            "ibm-granite/granite-4.0-micro",
            "ibm-granite/granite-4.0-micro-base",
            "ibm-granite/granite-4.0-h-micro",
            "ibm-granite/granite-4.0-h-micro-base",
            "ibm-granite/granite-4.0-h-tiny",
            "ibm-granite/granite-4.0-h-tiny-base",
            "ibm-granite/granite-4.0-h-small",
            "ibm-granite/granite-4.0-h-small-base",
            "ibm-granite/granite-4.0-tiny-preview",
            "ibm-granite/granite-4.0-tiny-base-preview",
        ],
        "RHOAI (vLLM)": ["meta-llama/Llama-3.2-1B-Instruct"],
        "RHAIIS (vLLM)": ["meta-llama/Llama-3.1-8B-Instruct"],
        "RHEL AI (vLLM)": ["meta-llama/Llama-3.1-8B-Instruct"],
    }
    return random.choice(models.get(provider, ["foo"]))


def generate_topic_summary() -> str:
    """Generate topic summary."""
    yaps = [
        [
            "Soudruzi,",
            "Na druhe strane",
            "Stejne tak",
            "Nesmime vsak zapominat, ze",
            "Timto zpusobem",
            "Zavaznost techto problemu je natolik zrejma, ze",
            "Kazdodenni praxe nam potvrzuje, ze",
            "Pestre a bohate zkusenosti",
            "Poslani organizace, zejmena pak",
            "Ideove uvahy nejvyssiho radu a rovnez",
        ],
        [
            "realizace planovanych vytycenych ukolu",
            "ramec a mista vychovy kadru",
            "stabilni a kvantitativni vzrust a sfera nasi aktivity",
            "vytvorena struktura organizace",
            "novy model organizacni cinnosti",
            "stale, informacne-propagandisticke zabezpeceni nasi prace",
            "dalsi rozvoj ruznych forem cinnosti",
            "upresneni a rozvoj struktur",
            "konzultace se sirokym aktivem",
            "pocatek kazdodenni prace na poli formovani pozice",
        ],
        [
            "hraje zavaznou roli pri utvareni",
            "vyzaduji od nas analyzy",
            "vyzaduji nalezeni a jednoznacne upresneni",
            "napomaha priprave a realizaci",
            "zabezpecuje sirokemu okruhu specialistu ucast pri tvorbe",
            "ve znacne mire podminuje vytvoreni",
            "umoznuje splnit vyznamne ukoly na rozpracovani",
            "umoznuje zhodnotit vyznam",
            "predstavuje pozoruhodny experiment proverky",
            "vyvolava proces zavadeni a modernizace",
        ],
        [
            "existujicich financnich a administrativnich podminek",
            "dalsich smeru rozvoje",
            "systemu masove ucasti",
            "pozic jednotlivych ucastniku k zadanym ukolum",
            "novych navrhu",
            "systemu vychovy kadru odpovidajicich aktualnim potrebam",
            "smeru progresivniho rozvoje",
            "odpovidajicich podminek aktivizace",
            "modelu rozvoje",
            "forem pusobeni",
        ],
    ]

    summary = " ".join([random.choice(yap) for yap in yaps]) + "."
    return summary


def store_new_user_conversation(session: Session, id: Optional[str] = None) -> None:
    """Store the new user conversation into database."""
    provider = generate_provider()
    model = generate_model_for_provider(provider)
    topic_summary = generate_topic_summary()
    conversation = UserConversation(
        id=id or get_suid(),
        user_id=get_suid(),
        last_used_model=model,
        last_used_provider=provider,
        topic_summary=topic_summary,
        last_message_at=datetime.now(UTC),
        message_count=1,
    )
    session.add(conversation)
    session.commit()


def update_user_conversation(session: Session, id: str) -> None:
    """Update existing conversation in the database."""
    provider = generate_provider()
    model = generate_model_for_provider(provider)
    topic_summary = generate_topic_summary()

    existing_conversation = session.query(UserConversation).filter_by(id=id).first()
    assert existing_conversation is not None

    existing_conversation.last_used_model = model
    existing_conversation.last_used_provider = provider
    existing_conversation.last_message_at = datetime.now(UTC)
    existing_conversation.message_count += 1
    existing_conversation.topic_summary = topic_summary
    session.commit()


def list_conversation_for_all_users(session: Session) -> None:
    """List all user conversations from the database."""
    query = session.query(UserConversation)

    user_conversations = query.all()
    assert user_conversations is not None
    assert len(user_conversations) >= 0


def benchmark_store_new_user_conversations(
    benchmark: BenchmarkFixture, records_to_insert: int
) -> None:
    with get_session() as session:
        # store bunch of conversations first
        for id in range(records_to_insert):
            store_new_user_conversation(session, str(id))
        # then perform the benchmark
        benchmark(store_new_user_conversation, session)


def _test_store_new_user_conversations_small_db(
    sqlite_database: None, benchmark: BenchmarkFixture
) -> None:
    """Benchmark for the DB operation to create and store new topic and conversation ID mapping."""
    benchmark_store_new_user_conversations(benchmark, 0)


def _test_store_new_user_conversations_middle_db(
    sqlite_database: None, benchmark: BenchmarkFixture
) -> None:
    """Benchmark for the DB operation to create and store new topic and conversation ID mapping."""
    benchmark_store_new_user_conversations(benchmark, MIDDLE_DB_RECORDS_COUNT)


def _test_store_new_user_conversations_large_db(
    sqlite_database: None, benchmark: BenchmarkFixture
) -> None:
    """Benchmark for the DB operation to create and store new topic and conversation ID mapping."""
    benchmark_store_new_user_conversations(benchmark, LARGE_DB_RECORDS_COUNT)


def benchmark_update_user_conversation(
    benchmark: BenchmarkFixture, records_to_insert: int
) -> None:
    with get_session() as session:
        # store bunch of conversations first
        # Ensure record "1234" exists for the update benchmark.
        # if records_to_insert <= 1234, range() won't include 1234, so insert it explicitly.
        if records_to_insert <= 1234:
            store_new_user_conversation(session, "1234")

        # pre-populate database with records
        for id in range(records_to_insert):
            store_new_user_conversation(session, str(id))

        # then perform the benchmark
        benchmark(update_user_conversation, session, "1234")


def test_update_user_conversation_small_db(
    sqlite_database: None,
    benchmark: BenchmarkFixture,
) -> None:
    """Benchmark for the DB operation to update existing conversation."""
    benchmark_update_user_conversation(benchmark, 0)


def test_update_user_conversation_middle_db(
    sqlite_database: None,
    benchmark: BenchmarkFixture,
) -> None:
    """Benchmark for the DB operation to update existing conversation."""
    benchmark_update_user_conversation(benchmark, MIDDLE_DB_RECORDS_COUNT)


def test_update_user_conversation_large_db(
    sqlite_database: None,
    benchmark: BenchmarkFixture,
) -> None:
    """Benchmark for the DB operation to update existing conversation."""
    benchmark_update_user_conversation(benchmark, LARGE_DB_RECORDS_COUNT)


def benchmark_list_conversations_for_all_users(
    benchmark: BenchmarkFixture, records_to_insert: int
) -> None:
    with get_session() as session:
        # store bunch of conversations first
        for id in range(records_to_insert):
            store_new_user_conversation(session, str(id))
        # then perform the benchmark
        benchmark(list_conversation_for_all_users, session)


def test_list_conversations_for_all_users_small_db(
    sqlite_database: None,
    benchmark: BenchmarkFixture
) -> None:
    """Benchmark for the DB operation to list all conversations."""
    benchmark_list_conversations_for_all_users(benchmark, 0)


def test_list_conversations_for_all_users_middle_db(
    sqlite_database: None,
    benchmark: BenchmarkFixture
) -> None:
    """Benchmark for the DB operation to list all conversations."""
    benchmark_list_conversations_for_all_users(benchmark, MIDDLE_DB_RECORDS_COUNT)


def test_list_conversations_for_all_users_large_db(
    sqlite_database: None,
    benchmark: BenchmarkFixture
) -> None:
    """Benchmark for the DB operation to list all conversations."""
    benchmark_list_conversations_for_all_users(benchmark, LARGE_DB_RECORDS_COUNT)

